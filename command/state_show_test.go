package command

import (
	"strings"
	"testing"

	"github.com/truecar-ops/terraform/addrs"
	"github.com/truecar-ops/terraform/configs/configschema"
	"github.com/truecar-ops/terraform/providers"
	"github.com/truecar-ops/terraform/states"
	"github.com/truecar-ops/terraform/terraform"
	"github.com/mitchellh/cli"
	"github.com/zclconf/go-cty/cty"
)

func TestStateShow(t *testing.T) {
	state := states.BuildState(func(s *states.SyncState) {
		s.SetResourceInstanceCurrent(
			addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: "test_instance",
				Name: "foo",
			}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance),
			&states.ResourceInstanceObjectSrc{
				AttrsJSON: []byte(`{"id":"bar","foo":"value","bar":"value"}`),
				Status:    states.ObjectReady,
			},
			addrs.AbsProviderConfig{
				Provider: addrs.NewDefaultProvider("test"),
				Module:   addrs.RootModule,
			},
		)
	})
	statePath := testStateFile(t, state)

	p := testProvider()
	p.GetSchemaReturn = &terraform.ProviderSchema{
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"foo": {Type: cty.String, Optional: true},
					"bar": {Type: cty.String, Optional: true},
				},
			},
		},
	}

	ui := new(cli.MockUi)
	c := &StateShowCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(p),
			Ui:               ui,
		},
	}

	args := []string{
		"-state", statePath,
		"test_instance.foo",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test that outputs were displayed
	expected := strings.TrimSpace(testStateShowOutput) + "\n"
	actual := ui.OutputWriter.String()
	if actual != expected {
		t.Fatalf("Expected:\n%q\n\nTo equal:\n%q", actual, expected)
	}
}

func TestStateShow_multi(t *testing.T) {
	submod, _ := addrs.ParseModuleInstanceStr("module.sub")
	state := states.BuildState(func(s *states.SyncState) {
		s.SetResourceInstanceCurrent(
			addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: "test_instance",
				Name: "foo",
			}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance),
			&states.ResourceInstanceObjectSrc{
				AttrsJSON: []byte(`{"id":"bar","foo":"value","bar":"value"}`),
				Status:    states.ObjectReady,
			},
			addrs.AbsProviderConfig{
				Provider: addrs.NewDefaultProvider("test"),
				Module:   addrs.RootModule,
			},
		)
		s.SetResourceInstanceCurrent(
			addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: "test_instance",
				Name: "foo",
			}.Instance(addrs.NoKey).Absolute(submod),
			&states.ResourceInstanceObjectSrc{
				AttrsJSON: []byte(`{"id":"foo","foo":"value","bar":"value"}`),
				Status:    states.ObjectReady,
			},
			addrs.AbsProviderConfig{
				Provider: addrs.NewDefaultProvider("test"),
				Module:   submod.Module(),
			},
		)
	})
	statePath := testStateFile(t, state)

	p := testProvider()
	p.GetSchemaReturn = &terraform.ProviderSchema{
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"foo": {Type: cty.String, Optional: true},
					"bar": {Type: cty.String, Optional: true},
				},
			},
		},
	}

	ui := new(cli.MockUi)
	c := &StateShowCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(p),
			Ui:               ui,
		},
	}

	args := []string{
		"-state", statePath,
		"test_instance.foo",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test that outputs were displayed
	expected := strings.TrimSpace(testStateShowOutput) + "\n"
	actual := ui.OutputWriter.String()
	if actual != expected {
		t.Fatalf("Expected:\n%q\n\nTo equal:\n%q", actual, expected)
	}
}

func TestStateShow_noState(t *testing.T) {
	tmp, cwd := testCwd(t)
	defer testFixCwd(t, tmp, cwd)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateShowCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(p),
			Ui:               ui,
		},
	}

	args := []string{
		"test_instance.foo",
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: %d", code)
	}
	if !strings.Contains(ui.ErrorWriter.String(), "No state file was found!") {
		t.Fatalf("expected a no state file error, got: %s", ui.ErrorWriter.String())
	}
}

func TestStateShow_emptyState(t *testing.T) {
	state := states.NewState()
	statePath := testStateFile(t, state)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateShowCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(p),
			Ui:               ui,
		},
	}

	args := []string{
		"-state", statePath,
		"test_instance.foo",
	}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: %d", code)
	}
	if !strings.Contains(ui.ErrorWriter.String(), "No instance found for the given address!") {
		t.Fatalf("expected a no instance found error, got: %s", ui.ErrorWriter.String())
	}
}

func TestStateShow_configured_provider(t *testing.T) {
	state := states.BuildState(func(s *states.SyncState) {
		s.SetResourceInstanceCurrent(
			addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: "test_instance",
				Name: "foo",
			}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance),
			&states.ResourceInstanceObjectSrc{
				AttrsJSON: []byte(`{"id":"bar","foo":"value","bar":"value"}`),
				Status:    states.ObjectReady,
			},
			addrs.AbsProviderConfig{
				Provider: addrs.NewDefaultProvider("test-beta"),
				Module:   addrs.RootModule,
			},
		)
	})
	statePath := testStateFile(t, state)

	p := testProvider()
	p.GetSchemaReturn = &terraform.ProviderSchema{
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"foo": {Type: cty.String, Optional: true},
					"bar": {Type: cty.String, Optional: true},
				},
			},
		},
	}

	ui := new(cli.MockUi)
	c := &StateShowCommand{
		Meta: Meta{
			testingOverrides: &testingOverrides{
				Providers: map[addrs.Provider]providers.Factory{
					addrs.NewDefaultProvider("test-beta"): providers.FactoryFixed(p),
				},
			},
			Ui: ui,
		},
	}

	args := []string{
		"-state", statePath,
		"test_instance.foo",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test that outputs were displayed
	expected := strings.TrimSpace(testStateShowOutput) + "\n"
	actual := ui.OutputWriter.String()
	if actual != expected {
		t.Fatalf("Expected:\n%q\n\nTo equal:\n%q", actual, expected)
	}
}

const testStateShowOutput = `
# test_instance.foo:
resource "test_instance" "foo" {
    bar = "value"
    foo = "value"
    id  = "bar"
}
`
