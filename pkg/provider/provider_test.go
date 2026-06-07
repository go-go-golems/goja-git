package provider

import (
	"encoding/json"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
)

func TestRegisterProvider(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	mod, ok := registry.ResolveModule(PackageID, ModuleName)
	if !ok {
		t.Fatalf("missing module %s.%s", PackageID, ModuleName)
	}
	if mod.DefaultAs != ModuleName {
		t.Fatalf("default alias = %q, want %q", mod.DefaultAs, ModuleName)
	}
}

func TestProviderRequiresAllowWrite(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	mod, ok := registry.ResolveModule(PackageID, ModuleName)
	if !ok {
		t.Fatalf("missing module")
	}
	if _, err := mod.New(providerapi.ModuleSetupContext{}); err == nil {
		t.Fatalf("expected missing allowWrite error")
	}
	if _, err := mod.New(providerapi.ModuleSetupContext{Config: json.RawMessage(`{"allowWrite": false}`)}); err == nil {
		t.Fatalf("expected allowWrite=false error")
	}
}

func TestModuleLoaderInstallsGitExports(t *testing.T) {
	registry := providerapi.NewRegistry()
	if err := Register(registry); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	mod, ok := registry.ResolveModule(PackageID, ModuleName)
	if !ok {
		t.Fatalf("missing module")
	}
	loader, err := mod.New(providerapi.ModuleSetupContext{
		Name:   ModuleName,
		As:     ModuleName,
		Config: json.RawMessage(`{"allowWrite": true}`),
	})
	if err != nil {
		t.Fatalf("create loader: %v", err)
	}

	vm := goja.New()
	moduleObj := vm.NewObject()
	exports := vm.NewObject()
	if err := moduleObj.Set("exports", exports); err != nil {
		t.Fatalf("set exports: %v", err)
	}
	loader(vm, moduleObj)
	if _, ok := goja.AssertFunction(exports.Get("open")); !ok {
		t.Fatalf("open export is not a function")
	}
	if _, ok := goja.AssertFunction(exports.Get("init")); !ok {
		t.Fatalf("init export is not a function")
	}
}
