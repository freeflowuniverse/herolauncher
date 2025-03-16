package vm

import (
	"fmt"
	"log"
)

// Manager handles VM operations
type Manager struct {
	vms map[string]*VM
}

// VM represents a virtual machine
type VM struct {
	Name        string
	CPU         int
	Memory      string
	Disks       []Disk
	Description string
	Running     bool
}

// Disk represents a VM disk
type Disk struct {
	Name string
	Size string
	Type string
}

// NewManager creates a new VM manager
func NewManager() *Manager {
	return &Manager{
		vms: make(map[string]*VM),
	}
}

// Define creates a new VM definition
func (m *Manager) Define(params Params) string {
	log.Printf("VM Manager: Define called with params: %v", params)
	
	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}
	
	// Check if VM already exists
	if _, exists := m.vms[name]; exists {
		return fmt.Sprintf("Error: VM '%s' already exists", name)
	}
	
	// Create new VM
	cpu := params.GetInt("cpu")
	if cpu <= 0 {
		cpu = 1
	}
	
	memory := params.Get("memory")
	if memory == "" {
		memory = "1GB"
	}
	
	description := params.Get("description")
	
	vm := &VM{
		Name:        name,
		CPU:         cpu,
		Memory:      memory,
		Description: description,
		Disks:       []Disk{},
		Running:     false,
	}
	
	// Add VM to map
	m.vms[name] = vm
	
	return fmt.Sprintf("VM '%s' defined successfully with %d CPU(s) and %s memory", 
		name, cpu, memory)
}

// Start starts a VM
func (m *Manager) Start(params Params) string {
	log.Printf("VM Manager: Start called with params: %v", params)
	
	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}
	
	// Find VM
	vm, exists := m.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}
	
	// Check if already running
	if vm.Running {
		return fmt.Sprintf("VM '%s' is already running", name)
	}
	
	// Start VM
	vm.Running = true
	return fmt.Sprintf("VM '%s' started successfully", name)
}

// Stop stops a VM
func (m *Manager) Stop(params Params) string {
	log.Printf("VM Manager: Stop called with params: %v", params)
	
	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}
	
	// Find VM
	vm, exists := m.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}
	
	// Check if already stopped
	if !vm.Running {
		return fmt.Sprintf("VM '%s' is already stopped", name)
	}
	
	// Stop VM
	vm.Running = false
	return fmt.Sprintf("VM '%s' stopped successfully", name)
}

// DiskAdd adds a disk to a VM
func (m *Manager) DiskAdd(params Params) string {
	log.Printf("VM Manager: DiskAdd called with params: %v", params)
	
	vmName := params.Get("name")
	if vmName == "" {
		return "Error: VM name is required"
	}
	
	// Find VM
	vm, exists := m.vms[vmName]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", vmName)
	}
	
	// Get disk parameters
	diskName := params.Get("disk_name")
	if diskName == "" {
		return "Error: Disk name is required"
	}
	
	// Check if disk already exists
	for _, disk := range vm.Disks {
		if disk.Name == diskName {
			return fmt.Sprintf("Error: Disk '%s' already exists on VM '%s'", diskName, vmName)
		}
	}
	
	// Create new disk
	size := params.Get("size")
	if size == "" {
		size = "10GB"
	}
	
	diskType := params.Get("type")
	if diskType == "" {
		diskType = "ssd"
	}
	
	disk := Disk{
		Name: diskName,
		Size: size,
		Type: diskType,
	}
	
	// Add disk to VM
	vm.Disks = append(vm.Disks, disk)
	
	return fmt.Sprintf("Disk '%s' added to VM '%s' with size %s and type %s", 
		diskName, vmName, size, diskType)
}

// Delete deletes a VM
func (m *Manager) Delete(params Params) string {
	log.Printf("VM Manager: Delete called with params: %v", params)
	
	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}
	
	// Find VM
	vm, exists := m.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}
	
	// Check if VM is running and force flag is not set
	if vm.Running && !params.GetBool("force") {
		return fmt.Sprintf("Error: VM '%s' is running. Use force:true to delete anyway", name)
	}
	
	// Delete VM
	delete(m.vms, name)
	return fmt.Sprintf("VM '%s' deleted successfully", name)
}

// List returns a list of all VMs
func (m *Manager) List() string {
	log.Printf("VM Manager: List called")
	
	if len(m.vms) == 0 {
		return "No VMs defined"
	}
	
	var result string
	result = "Defined VMs:\n"
	
	for _, vm := range m.vms {
		status := "stopped"
		if vm.Running {
			status = "running"
		}
		
		result += fmt.Sprintf("- %s (%s): %d CPU, %s memory\n", 
			vm.Name, status, vm.CPU, vm.Memory)
		
		if vm.Description != "" {
			result += fmt.Sprintf("  Description: %s\n", vm.Description)
		}
		
		if len(vm.Disks) > 0 {
			result += "  Disks:\n"
			for _, disk := range vm.Disks {
				result += fmt.Sprintf("  - %s: %s %s\n", 
					disk.Name, disk.Size, disk.Type)
			}
		}
	}
	
	return result
}

// Status returns the status of a VM
func (m *Manager) Status(params Params) string {
	log.Printf("VM Manager: Status called with params: %v", params)
	
	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}
	
	// Find VM
	vm, exists := m.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}
	
	// Get VM status
	status := "stopped"
	if vm.Running {
		status = "running"
	}
	
	result := fmt.Sprintf("VM '%s' status: %s\n", vm.Name, status)
	result += fmt.Sprintf("CPU: %d\n", vm.CPU)
	result += fmt.Sprintf("Memory: %s\n", vm.Memory)
	
	if vm.Description != "" {
		result += fmt.Sprintf("Description: %s\n", vm.Description)
	}
	
	if len(vm.Disks) > 0 {
		result += "Disks:\n"
		for _, disk := range vm.Disks {
			result += fmt.Sprintf("- %s: %s %s\n", 
				disk.Name, disk.Size, disk.Type)
		}
	}
	
	return result
}
