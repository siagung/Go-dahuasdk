package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	TRUE  = 1
	FALSE = 0
	NULL  = 0
)

var (
	//dhnetsdkDll *syscall.DLL
	dhnetsdkDll            *syscall.LazyDLL
	client_Init            *syscall.LazyProc
	client_SetNetworkParam *syscall.LazyProc
	client_LoginEx2        *syscall.LazyProc
	client_Logout          *syscall.LazyProc
	client_Cleanup         *syscall.LazyProc
	client_SetDevConfig    *syscall.LazyProc
	client_GetDevConfig    *syscall.LazyProc
	client_GetLastError    *syscall.LazyProc
	client_QuerySystemInfo *syscall.LazyProc
)

type DH_DEV_ENABLE_INFO struct {
	IsFucEnable [512]uint32
}

func init() {
	//dhnetsdkDll = syscall.MustLoadDLL("sdk/dhnetsdk.dll")
	dhnetsdkDll = syscall.NewLazyDLL("sdk/dhnetsdk.dll")

	client_Init = dhnetsdkDll.NewProc("CLIENT_Init")
	client_SetNetworkParam = dhnetsdkDll.NewProc("CLIENT_SetNetworkParam")
	client_LoginEx2 = dhnetsdkDll.NewProc("CLIENT_LoginEx2")
	client_Logout = dhnetsdkDll.NewProc("CLIENT_Logout")
	client_Cleanup = dhnetsdkDll.NewProc("CLIENT_Cleanup")
	client_SetDevConfig = dhnetsdkDll.NewProc("CLIENT_SetDevConfig")
	client_GetDevConfig = dhnetsdkDll.NewProc("CLIENT_GetDevConfig")
	client_GetLastError = dhnetsdkDll.NewProc("CLIENT_GetLastError")
	client_QuerySystemInfo = dhnetsdkDll.NewProc("CLIENT_QuerySystemInfo")
}

func main() {
	ok := CLIENT_Init(0, 0)
	if !ok {
		fmt.Println("Failed to initialize dll.")
		return
	}

	pNetParam := new(NET_PARAM)
	pNetParam.nConnectTryNum = 2
	pNetParam.nGetDevInfoTime = 3000
	CLIENT_SetNetworkParam(pNetParam)
	ip := "192.168.1.108"
	pchDVRIP := StringToBytePtr(ip)
	pchUserName := StringToBytePtr("admin")
	pchPassword := StringToBytePtr("Admin123")
	var err int32
	var deviceInfo NET_DEVICEINFO_Ex
	lLoginID := CLIENT_LoginEx2(pchDVRIP, 37777, pchUserName, pchPassword, EM_LOGIN_SPEC_CAP_TCP, 0, &deviceInfo, &err)
	if lLoginID == 0 {
		fmt.Printf("Login to camera [%s] failed.%v\n", ip, err)
		return
	}
	fmt.Printf("Login to camera [%s] succeeded, login ID: %d.\n", ip, lLoginID)

	var pSysInfoBuffer DH_DEV_ENABLE_INFO
	var nSysInfolen int32
	ok = CLIENT_QuerySystemInfo(lLoginID, ABILITY_DEVALL_INFO, (*byte)(unsafe.Pointer(&pSysInfoBuffer)), int32(unsafe.Sizeof(pSysInfoBuffer)), &nSysInfolen, 1000)
	if !ok {
		fmt.Printf("Failed to get the list of supported features for camera [%s].", ip)
		return
	}
	fmt.Printf("Obtaining the list of supported functions for camera [%s] succeeded.\n", ip)

	fmt.Printf("Camera [%s] supports the following functions:\n", ip)
	for index, name := range names {
		if pSysInfoBuffer.IsFucEnable[index] != FALSE {
			fmt.Printf("-- %s\n", name)
		}
	}
}

func StringToBytePtr(str string) *byte {
	p, err := syscall.BytePtrFromString(str)
	if err != nil {
		return nil
	}
	return p
	//return windows.BytePtrToString([]byte(str))
}

func CLIENT_Init(cbDisConnect uintptr, dwUser uint32) bool {
	ret, _, _ := client_Init.Call(cbDisConnect, uintptr(dwUser))

	return ret == TRUE
}

func CLIENT_SetNetworkParam(pNetParam *NET_PARAM) {
	client_SetNetworkParam.Call(uintptr(unsafe.Pointer(pNetParam)))
}

func CLIENT_LoginEx2(pchDVRIP *byte, wDVRPort uint16, pchUserName, pchPassword *byte,
	emSpecCap EM_LOGIN_SPAC_CAP_TYPE, pCapParam uintptr, lpDeviceInfo *NET_DEVICEINFO_Ex, err *int32) int64 {
	ret, _, _ := client_LoginEx2.Call(uintptr(unsafe.Pointer(pchDVRIP)),
		uintptr(wDVRPort),
		uintptr(unsafe.Pointer(pchUserName)),
		uintptr(unsafe.Pointer(pchPassword)),
		uintptr(emSpecCap),
		pCapParam,
		uintptr(unsafe.Pointer(lpDeviceInfo)),
		uintptr(unsafe.Pointer(err)))

	return int64(ret)
}

func CLIENT_SetDevConfig(lLoginID int64, dwCommand uint32, lChannel int32, lpInBuffer uintptr, dwInBufferSize uint32, waittime int32) bool {
	ret, _, _ := client_SetDevConfig.Call(uintptr(lLoginID),
		uintptr(dwCommand),
		uintptr(lChannel),
		lpInBuffer,
		uintptr(dwInBufferSize),
		uintptr(waittime))

	return ret == TRUE
}

func CLIENT_GetDevConfig(lLoginID int64, dwCommand uint32, lChannel int32, lpOutBuffer uintptr, dwOutBufferSize uint32, lpBytesReturned *uint32, waittime int32) bool {
	ret, _, _ := client_SetDevConfig.Call(uintptr(lLoginID),
		uintptr(dwCommand),
		uintptr(lChannel),
		lpOutBuffer,
		uintptr(dwOutBufferSize),
		uintptr(unsafe.Pointer(lpBytesReturned)),
		uintptr(waittime))

	return ret == TRUE
}

func CLIENT_QuerySystemInfo(lLoginID int64, nSystemType int32, pSysInfoBuffer *byte, maxlen int32, nSysInfolen *int32, waittime int) bool {
	ret, _, _ := client_QuerySystemInfo.Call(uintptr(lLoginID),
		uintptr(nSystemType),
		uintptr(unsafe.Pointer(pSysInfoBuffer)),
		uintptr(maxlen),
		uintptr(unsafe.Pointer(nSysInfolen)),
		uintptr(waittime))

	return ret == TRUE
}

func CLIENT_GetLastError() uint32 {
	ret, _, _ := client_GetLastError.Call()

	return uint32(ret)
}

// var netSnmpCFG DHDEV_NET_SNMP_CFG
// netSnmpCFG.bEnable = '1'
// netSnmpCFG.bSNMPV1 = '1'
// netSnmpCFG.bSNMPV2 = '1'
// netSnmpCFG.iSNMPPort = 161
// netSnmpCFG.iTrapPort = 162
// var public [DH_MAX_SNMP_COMMON_LEN]byte
// copy(public[:], "public")

// var privite [DH_MAX_SNMP_COMMON_LEN]byte
// copy(public[:], "privite")
// netSnmpCFG.szReadCommon = public
// netSnmpCFG.szWriteCommon = privite

// ok = CLIENT_SetDevConfig(lLoginID, DH_DEV_SNMP_CFG, -1, uintptr(unsafe.Pointer(&netSnmpCFG)), uint32(unsafe.Sizeof(netSnmpCFG)), 3000)

// if ok {
//  fmt.Printf("设置 %s SNMP成功: %v\n", ip, ok)
// } else {
//  fmt.Printf("设置 %s SNMP失败: %v,ID: %v\n", ip, ok, lLoginID)
// }

// var curDateTime NET_TIME = NET_TIME{2016, 12, 1, 10, 00, 00}
// ok = CLIENT_SetDevConfig(lLoginID, DH_DEV_TIMECFG, -1, uintptr(unsafe.Pointer(&curDateTime)), uint32(unsafe.Sizeof(curDateTime)), 1000)
// if ok {
//  fmt.Println("设置成功")
// } else {
//  fmt.Println("设置失败.")
// }

// var dwRet uint32 = 0

// ok = CLIENT_GetDevConfig(lLoginID, DH_DEV_TIMECFG, -1, uintptr(unsafe.Pointer(&curDateTime)), uint32(unsafe.Sizeof(curDateTime)), &dwRet, 3000)
// if ok {
//  fmt.Println("get 成功", ok)
// }
// fmt.Println(curDateTime.dwYear, dwRet)
