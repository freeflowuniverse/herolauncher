package main

import (
	"fmt"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
)

// VMHandler handles VM-related actions
type VMHandler struct {
	handlerfactory.BaseHandler
	vms map[string]*VM
}

// VM represents a virtual machine
type VM struct {
	Name        string
	CPU         int
	Memory      string
	Storage     string
	Description string
	Running     bool
	Disks       []Disk
}

// Disk represents a disk attached to a VM
type Disk struct {
	Size string
	Type string
}

// NewVMHandler creates a new VM handler
func NewVMHandler() *VMHandler {
	return &VMHandler{
		BaseHandler: handlerfactory.BaseHandler{
			ActorName: "vm",
		},
		vms: make(map[string]*VM),
	}
}

// Define handles the vm.define action
func (h *VMHandler) Define(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Check if VM already exists
	if _, exists := h.vms[name]; exists {
		return fmt.Sprintf("Error: VM '%s' already exists", name)
	}

	// Create new VM
	cpu := params.GetIntDefault("cpu", 1)
	memory := params.Get("memory")
	if memory == "" {
		memory = "1GB"
	}
	storage := params.Get("storage")
	if storage == "" {
		storage = "10GB"
	}
	description := params.Get("description")

	vm := &VM{
		Name:        name,
		CPU:         cpu,
		Memory:      memory,
		Storage:     storage,
		Description: description,
		Running:     false,
		Disks:       []Disk{},
	}

	// Add VM to map
	h.vms[name] = vm

	return fmt.Sprintf("VM '%s' defined successfully with %d CPU, %s memory, and %s storage", 
		name, cpu, memory, storage)
}

// Start handles the vm.start action
func (h *VMHandler) Start(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Find VM
	vm, exists := h.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}

	// Start VM
	if vm.Running {
		return fmt.Sprintf("VM '%s' is already running", name)
	}

	vm.Running = true
	return fmt.Sprintf("VM '%s' started successfully", name)
}

// Stop handles the vm.stop action
func (h *VMHandler) Stop(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Find VM
	vm, exists := h.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}

	// Stop VM
	if !vm.Running {
		return fmt.Sprintf("VM '%s' is already stopped", name)
	}

	vm.Running = false
	return fmt.Sprintf("VM '%s' stopped successfully", name)
}

// DiskAdd handles the vm.disk_add action
func (h *VMHandler) DiskAdd(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Find VM
	vm, exists := h.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}

	// Add disk
	size := params.Get("size")
	if size == "" {
		size = "10GB"
	}
	diskType := params.Get("type")
	if diskType == "" {
		diskType = "HDD"
	}

	disk := Disk{
		Size: size,
		Type: diskType,
	}

	vm.Disks = append(vm.Disks, disk)
	return fmt.Sprintf("Added %s %s disk to VM '%s'", size, diskType, name)
}

// Delete handles the vm.delete action
func (h *VMHandler) Delete(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Find VM
	vm, exists := h.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}

	// Check if VM is running and force flag is not set
	if vm.Running && !params.GetBool("force") {
		return fmt.Sprintf("Error: VM '%s' is running. Use force:true to delete anyway", name)
	}

	// Delete VM
	delete(h.vms, name)
	return fmt.Sprintf("VM '%s' deleted successfully", name)
}

// List handles the vm.list action
func (h *VMHandler) List(script string) string {
	if len(h.vms) == 0 {
		return "No VMs defined"
	}

	var result strings.Builder
	result.WriteString("Defined VMs:\n")

	for _, vm := range h.vms {
		status := "stopped"
		if vm.Running {
			status = "running"
		}

		result.WriteString(fmt.Sprintf("- %s (%s): %d CPU, %s memory, %s storage\n", 
			vm.Name, status, vm.CPU, vm.Memory, vm.Storage))
		
		if len(vm.Disks) > 0 {
			result.WriteString("  Attached disks:\n")
			for i, disk := range vm.Disks {
				result.WriteString(fmt.Sprintf("  %d. %s %s\n", i+1, disk.Size, disk.Type))
			}
		}
	}

	return result.String()
}

// Status handles the vm.status action
func (h *VMHandler) Status(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: VM name is required"
	}

	// Find VM
	vm, exists := h.vms[name]
	if !exists {
		return fmt.Sprintf("Error: VM '%s' not found", name)
	}

	// Return VM status
	status := "stopped"
	if vm.Running {
		status = "running"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("VM '%s' status:\n", name))
	result.WriteString(fmt.Sprintf("- Status: %s\n", status))
	result.WriteString(fmt.Sprintf("- CPU: %d\n", vm.CPU))
	result.WriteString(fmt.Sprintf("- Memory: %s\n", vm.Memory))
	result.WriteString(fmt.Sprintf("- Storage: %s\n", vm.Storage))
	
	if vm.Description != "" {
		result.WriteString(fmt.Sprintf("- Description: %s\n", vm.Description))
	}
	
	if len(vm.Disks) > 0 {
		result.WriteString("- Attached disks:\n")
		for i, disk := range vm.Disks {
			result.WriteString(fmt.Sprintf("  %d. %s %s\n", i+1, disk.Size, disk.Type))
		}
	}

	return result.String()
}
