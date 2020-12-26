package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	cgCPUPathPrefix    = "/sys/fs/cgroup/cpu/gobox/"
	cgPidPathPrefix    = "/sys/fs/cgroup/pids/gobox/"
	cgMemoryPathPrefix = "/sys/fs/cgroup/memory/gobox/"
)

func initCGroups(PID, containerID, lmPIDs, lmCfsQuotaUs, lmMemory string) error {
	_, _ = os.Stderr.WriteString(fmt.Sprintf("[Init CGroups](%s, %s, %s, %s, %s) start...\n", PID, containerID, lmPIDs, lmCfsQuotaUs, lmMemory))

	dirs := []string{
		filepath.Join(cgCPUPathPrefix, containerID),
		filepath.Join(cgPidPathPrefix, containerID),
		filepath.Join(cgMemoryPathPrefix, containerID),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("os.MkdirAll(%s, os.ModePerm) failed, err: %s\n", dir, err.Error()))
			return err
		}
		if err := ioutil.WriteFile(dir+"/notify_on_release", [](byte)("1"), 0700); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("ioutil.WriteFile(%s, 1, 0700) failed, err: %s\n", dir+"/notify_on_release", err.Error()))
			return err
		}
		if err := ioutil.WriteFile(dir+"/cgroup.procs", [](byte)(PID), 0700); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("ioutil.WriteFile(%s, %s, 0700) failed, err: %s\n", dir+"/cgroup.procs", PID, err.Error()))
			return err
		}
	}

	if err := cpuCGroup(PID, containerID, lmCfsQuotaUs); err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("cpuCGroup(%s, %s, %s) failed, err: %s\n", PID, containerID, lmCfsQuotaUs, err.Error()))
		return err
	}

	if err := pidCGroup(PID, containerID, lmPIDs); err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("pidCGroup(%s, %s, %s) failed, err: %s\n", PID, containerID, lmPIDs, err.Error()))
		return err
	}

	if err := memoryCGroup(PID, containerID, lmMemory); err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("memoryCGroup(%s, %s, %s) failed, err: %s\n", PID, containerID, lmMemory, err.Error()))
		return err
	}

	_, _ = os.Stderr.WriteString(fmt.Sprintf("[Init CGroups](%s, %s, %s, %s, %s) done...\n", PID, containerID, lmPIDs, lmCfsQuotaUs, lmMemory))
	return nil
}

func cpuCGroup(PID, containerID, lmCfsQuotaUs string) error {
	cgCPUPath := filepath.Join(cgCPUPathPrefix, containerID)
	mapping := map[string]string{
		"tasks":            PID,
		"cpu.cfs_quota_us": lmCfsQuotaUs,
	}

	for k, v := range mapping {
		p := filepath.Join(cgCPUPath, k)

		if err := ioutil.WriteFile(p, []byte(v), 0644); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Writing [%s] to file: %s failed\n", v, p))
			return err
		}
		c, _ := ioutil.ReadFile(p)
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Content of %s is: %s", p, c))
	}

	return nil
}

func pidCGroup(PID, containerID, lmPIDs string) error {
	cgPidPath := filepath.Join(cgPidPathPrefix, containerID)
	mapping := map[string]string{
		"pids.max": lmPIDs,
	}

	for k, v := range mapping {
		p := filepath.Join(cgPidPath, k)

		if err := ioutil.WriteFile(p, []byte(v), 0644); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Writing [%s] to file: %s failed\n", v, p))
			return err
		}
		c, _ := ioutil.ReadFile(p)
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Content of %s is: %s", p, c))
	}

	return nil
}

func memoryCGroup(PID, containerID, lmMemory string) error {
	cgMemoryPath := filepath.Join(cgMemoryPathPrefix, containerID)
	mapping := map[string]string{
		"memory.kmem.limit_in_bytes": "64m",
		"tasks":                      PID,
		"memory.limit_in_bytes":      fmt.Sprintf("%sm", lmMemory),
	}

	for key, value := range mapping {
		path := filepath.Join(cgMemoryPath, key)
		if err := ioutil.WriteFile(path, []byte(value), 0644); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Writing [%s] to file: %s failed\n", value, path))
			return err
		}
		c, _ := ioutil.ReadFile(path)
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Content of %s is: %s", path, c))
	}

	return nil
}
