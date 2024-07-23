package dcgm

import "C"
import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// 初始化rocm_smi
func Init() error {
	return rsmiInit()
}

// 关闭rocm_smi
func ShutDown() error {
	return rsmiShutdown()
}

// 获取GPU数量
func NumMonitorDevices() (int, error) {
	return rsmiNumMonitorDevices()
}

// 获取设备利用率计数器
func UtilizationCount(dvInd int, utilizationCounters []RSMIUtilizationCounter, count int) (timestamp int64, err error) {
	return rsmiUtilizationCountGet(dvInd, utilizationCounters, count)
}

// 获取设备名称
func DevName(dvInd int) (name string, err error) {
	return rsmiDevNameGet(dvInd)
}

// 获取设备sku
func DevSku(dvInd int) (sku int, err error) {
	return rsmiDevSkuGet(dvInd)
}

// 获取设备品牌名称
func DevBrand(dvInd int) (brand string, err error) {
	return rsmiDevBrandGet(dvInd)
}

// 获取设备供应商名称
func DevVendorName(dvInd int) string {
	return rsmiDevVendorNameGet(dvInd)
}

// 获取设备显存供应商名称
func DevVramVendor(dvInd int) string {
	return rsmiDevVramVendorGet(dvInd)
}

// 获取可用的pcie带宽列表
func DevPciBandwidth(dvInd int) RSMIPcieBandwidth {
	return rsmiDevPciBandwidthGet(dvInd)

}

// 内存使用百分比
func MemoryPercent(dvInd int) int {
	return rsmiDevMemoryBusyPercentGet(dvInd)
}

// 获取设备温度值
//func DevTemp(dvInd int) int64 {
//	return go_rsmi_dev_temp_metric_get(dvInd)
//}

// 获取设别性能级别
func DevPerfLevel(dvInd int) (perf RSMIDevPerfLevel, err error) {
	return rsmiDevPerfLevelGet(dvInd)
}

// 设置设备PowerPlay性能级别
func DevPerfLevelSet(dvInd int, level RSMIDevPerfLevel) error {
	return rsmiDevPerfLevelSet(dvInd, level)
}

// 获取gpu度量信息
func DevGpuMetricsInfo(dvInd int) (gpuMetrics RSMIGPUMetrics, err error) {
	return rsmiDevGpuMetricsInfoGet(dvInd)
}

// 获取设备监控中的指标
func CollectDeviceMetrics() (monitorInfos []MonitorInfo, err error) {
	numMonitorDevices, err := rsmiNumMonitorDevices()
	if err != nil {
		return nil, err
	}
	for i := 0; i < numMonitorDevices; i++ {
		bdfid, err := rsmiDevPciIdGet(i)
		if err != nil {
			return nil, err
		}
		// 解析BDFID
		domain := (bdfid >> 32) & 0xffffffff
		bus := (bdfid >> 8) & 0xff
		dev := (bdfid >> 3) & 0x1f
		function := bdfid & 0x7
		// 格式化PCI ID
		pciBusNumber := fmt.Sprintf("%04x:%02x:%02x.%x", domain, bus, dev, function)
		//设备序列号
		deviceId := rsmiDevSerialNumberGet(i)
		//获取设备类型标识id
		devTypeId, _ := rsmiDevIdGet(i)
		//型号名称
		devTypeName := type2name[fmt.Sprintf("%x", devTypeId)]
		//设备温度
		temperature := rsmiDevTempMetricGet(i, 0, RSMI_TEMP_CURRENT)
		t, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(temperature)/1000.0), 64)
		if err != nil {
			return nil, err
		}
		//设备平均功耗
		powerUsage := rsmiDevPowerAveGet(i, 0)
		pu, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(powerUsage)/1000000.0), 64)
		glog.Info("🔋 DCU[%v] power cap : %v \n", i, pu)
		//获取设备功率上限
		powerCap, _ := rsmiDevPowerCapGet(i, 0)
		pc, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(powerCap)/1000000.0), 64)
		glog.Info("\U0001FAAB DCU[%v] power usage : %v \n", i, pc)
		//获取设备内存总量
		memoryCap, _ := rsmiDevMemoryTotalGet(i, RSMI_MEM_TYPE_FIRST)
		mc, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryCap)/1.0), 64)
		glog.Info(" DCU[%v] memory cap : %v \n", i, mc)
		//获取设备内存使用量
		memoryUsed, _ := rsmiDevMemoryUsageGet(i, RSMI_MEM_TYPE_FIRST)
		mu, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryUsed)/1.0), 64)
		glog.Info(" DCU[%v] memory used : %v \n", i, mu)
		//获取设备设备忙碌时间百分比
		utilizationRate, _ := rsmiDevBusyPercentGet(i)
		ur, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(utilizationRate)/1.0), 64)
		glog.Info(" DCU[%v] utilization rate : %v \n", i, ur)
		//获取pcie流量信息
		sent, received, maxPktSz := rsmiDevPciThroughputGet(i)
		pcieBwMb, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(received+sent)*float64(maxPktSz)/1024.0/1024.0), 64)
		glog.Info(" DCU[%v] PCIE  bandwidth : %v \n", i, pcieBwMb)
		//获取设备系统时钟速度列表
		clk, _ := rsmiDevGpuClkFreqGet(i, RSMI_CLK_TYPE_SYS)
		sclk, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(clk.Frequency[clk.Current])/1000000.0), 64)
		glog.Info(" DCU[%v] SCLK : %v \n", i, sclk)
		monitorInfo := MonitorInfo{
			PicBusNumber:    pciBusNumber,
			DeviceId:        deviceId,
			SubSystemName:   devTypeName,
			Temperature:     t,
			PowerUsage:      pu,
			powerCap:        pc,
			MemoryCap:       mc,
			MemoryUsed:      mu,
			UtilizationRate: ur,
			PcieBwMb:        pcieBwMb,
			Clk:             sclk,
		}
		monitorInfos = append(monitorInfos, monitorInfo)
	}
	glog.Info("monitorInfos: ", dataToJson(monitorInfos))
	return
}

// 设备的总线
func PicBusInfo(dvInd int) (picID string, err error) {
	bdfid, err := rsmiDevPciIdGet(dvInd)
	if err != nil {
		return "", err
	}
	// Parse BDFID
	domain := (bdfid >> 32) & 0xffffffff
	bus := (bdfid >> 8) & 0xff
	devID := (bdfid >> 3) & 0x1f
	function := bdfid & 0x7
	// Format and return the bus identifier
	picID = fmt.Sprintf("%04X:%02X:%02X.%X", domain, bus, devID, function)
	return
}

// 获取风扇转速信息
func FanSpeedInfo(dvInd int) (fanLevel int64, fanPercentage float64, err error) {
	// 当前转速
	fanLevel, err = rsmiDevFanSpeedGet(dvInd, 0)
	if err != nil {
		return 0, 0, err
	}
	// 最大转速
	fanMax, err := rsmiDevFanSpeedMaxGet(dvInd, 0)
	if err != nil {
		return 0, 0, err
	}
	// Calculate fan speed percentage
	fanPercentage = (float64(fanLevel) / float64(fanMax)) * 100
	return
}

// 当前GPU使用的百分比
func GPUUse(dvInd int) (percent int, err error) {
	percent, err = rsmiDevBusyPercentGet(dvInd)
	if err != nil {
		return 0, err
	}
	return
}

// 设备ID的十六进制值
func rsmiDevIDGet(dvInd int) (id int, err error) {
	id, err = rsmiDevIdGet(dvInd)
	if err != nil {
		return 0, err
	}
	return
}

// 设备的最大功率
func MaxPower(dvInd int) (power int64, err error) {
	power, err = rsmiDevPowerCapGet(dvInd, 0)
	if err != nil {
		return 0, err
	}
	return (power / 1000000), nil
}

// 设备的指定内存使用情况 memType:[vram|vis_vram|gtt]
func MemInfo(dvInd int, memType string) (memUsed int64, memTotal int64, err error) {
	memType = strings.ToUpper(memType)
	if !contains(memoryTypeL, memType) {
		fmt.Println(dvInd, fmt.Sprintf("Invalid memory type %s", memType))
		return 0, 0, fmt.Errorf("invalid memory type")
	}
	memTypeIndex := RSMIMemoryType(indexOf(memoryTypeL, memType))
	memUsed, err = rsmiDevMemoryUsageGet(dvInd, memTypeIndex)
	if err != nil {
		return memUsed, memTotal, err
	}
	fmt.Println(dvInd, fmt.Sprintf("memUsed: %d", memUsed))
	memTotal, err = rsmiDevMemoryTotalGet(dvInd, memTypeIndex)
	if err != nil {
		return memUsed, memTotal, err
	}
	fmt.Println(dvInd, fmt.Sprintf("memTotal: %d", memTotal))
	return
}

// 获取设备信息列表
func DeviceInfos() (deviceInfos []DeviceInfo, err error) {
	numDevices, err := rsmiNumMonitorDevices()
	if err != nil {
		return nil, err
	}
	for i := 0; i < numDevices; i++ {
		bdfid, err := rsmiDevPciIdGet(i)
		if err != nil {
			return nil, err
		}
		// 解析BDFID
		domain := (bdfid >> 32) & 0xffffffff
		bus := (bdfid >> 8) & 0xff
		dev := (bdfid >> 3) & 0x1f
		function := bdfid & 0x7
		// 格式化PCI ID
		pciBusNumber := fmt.Sprintf("%04X:%02X:%02X.%X", domain, bus, dev, function)
		//设备序列号
		deviceId := rsmiDevSerialNumberGet(i)
		//获取设备类型标识id
		devTypeId, _ := rsmiDevIdGet(i)
		devType := fmt.Sprintf("%x", devTypeId)
		//型号名称
		devTypeName := type2name[devType]
		//获取设备内存总量
		memoryTotal, _ := rsmiDevMemoryTotalGet(i, RSMI_MEM_TYPE_FIRST)
		mt, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryTotal)/1.0), 64)
		glog.Info(" DCU[%v] memory total : %v \n", i, mt)
		//获取设备内存使用量
		memoryUsed, _ := rsmiDevMemoryUsageGet(i, RSMI_MEM_TYPE_FIRST)
		mu, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryUsed)/1.0), 64)
		glog.Info(" DCU[%v] memory used : %v \n", i, mu)
		computeUnit := computeUnitType[devTypeName]
		glog.Info(" DCU[%v] computeUnit : %v \n", i, computeUnit)
		deviceInfo := DeviceInfo{
			DvInd:        i,
			DeviceId:     deviceId,
			DevType:      devType,
			DevTypeName:  devTypeName,
			PicBusNumber: pciBusNumber,
			MemoryTotal:  mt,
			MemoryUsed:   mu,
			ComputeUnit:  computeUnit,
		}
		deviceInfos = append(deviceInfos, deviceInfo)
	}
	glog.Info("deviceInfos: ", dataToJson(deviceInfos))
	return
}

// pid的进程名
func ProcessName(pid int) string {
	if pid < 1 {
		glog.Info("PID must be greater than 0")
		return "UNKNOWN"
	}
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		glog.Info("Error executing command:", err)
		return "UNKNOWN"
	}
	pName := out.String()
	if pName == "" {
		return "UNKNOWN"
	}
	// Remove the substrings surrounding from process name (b' and \n')
	pName = strings.TrimPrefix(pName, "b'")
	pName = strings.TrimSuffix(pName, "\\n'")
	glog.Info("Process name: %s\n", pName)
	return strings.TrimSpace(pName)
}

// 设备的当前性能水平
func PerfLevel(dvInd int) (perf string, err error) {
	level, err := rsmiDevPerfLevelGet(dvInd)
	if err != nil {
		return perf, err
	}
	perf = perfLevelString(int(level))
	glog.Info("Perf level: %s\n", perf)
	return
}

// getPid 获取特定应用程序的进程 ID
func PidByName(name string) (pid string, err error) {
	glog.Info("pidName: %s\n", name)
	cmd := exec.Command("pidof", name)
	output, err := cmd.Output()
	glog.Info("output:", output)
	//if err != nil {
	//	return "", fmt.Errorf("error getting pid: %v", err)
	//}
	if err != nil {
		glog.Info("Error: %v\nOutput: %s", err, string(output))
	} else {
		glog.Info("Output: %s", string(output))
	}
	// 移除末尾的换行符并返回 PID
	pid = strings.TrimSpace(string(output))
	glog.Info("pid: %s\n", pid)
	return
}