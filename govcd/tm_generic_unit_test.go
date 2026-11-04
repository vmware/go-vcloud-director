//go:build unit || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package govcd

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

type tmTestEntity struct{ ID, Name string }

func Test_tmEntityByName(t *testing.T) {
	entities := []*tmTestEntity{{"1", "a"}, {"2", "b"}, {"3", "b"}}
	nameOf := func(e *tmTestEntity) string { return e.Name }

	found, err := tmEntityByName("test", "a", entities, nameOf)
	if err != nil || found.ID != "1" {
		t.Errorf("expected entity 1, got %v, %v", found, err)
	}
	if _, err = tmEntityByName("test", "b", entities, nameOf); err == nil || !strings.Contains(err.Error(), "more than one") {
		t.Errorf("expected a duplicate error, got %v", err)
	}
	if _, err = tmEntityByName("test", "c", entities, nameOf); !ContainsNotFound(err) {
		t.Errorf("expected a not found error, got %v", err)
	}
	if _, err = tmEntityByName("test", "", entities, nameOf); err == nil || ContainsNotFound(err) {
		t.Errorf("expected an empty name error, got %v", err)
	}
}

func Test_tmEntityByNameOrId(t *testing.T) {
	const urn = "urn:vcloud:test:5f0cfb9d-3d1b-4c44-8b52-8c0fb2d3a111"
	other := errors.New("boom")
	var idCalls, nameCalls int
	byName := func(name string) (*tmTestEntity, error) {
		nameCalls++
		return &tmTestEntity{"by-name", name}, nil
	}
	tests := []struct {
		name       string
		identifier string
		byId       func(string) (*tmTestEntity, error)
		wantId     string
		wantErr    error
		idCalls    int
		nameCalls  int
	}{
		{"URN found by ID", urn, func(id string) (*tmTestEntity, error) { return &tmTestEntity{id, "x"}, nil }, urn, nil, 1, 0},
		{"URN not found falls back to name", urn, func(string) (*tmTestEntity, error) { return nil, fmt.Errorf("%s: gone", ErrorEntityNotFound) }, "by-name", nil, 1, 1},
		{"URN with other error", urn, func(string) (*tmTestEntity, error) { return nil, other }, "", other, 1, 0},
		{"plain name skips ID lookup", "my-name", func(string) (*tmTestEntity, error) { return nil, other }, "by-name", nil, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idCalls, nameCalls = 0, 0
			byId := func(id string) (*tmTestEntity, error) {
				idCalls++
				return tt.byId(id)
			}
			got, err := tmEntityByNameOrId(tt.identifier, byId, byName)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if err == nil && got.ID != tt.wantId {
				t.Errorf("expected ID %s, got %s", tt.wantId, got.ID)
			}
			if idCalls != tt.idCalls || nameCalls != tt.nameCalls {
				t.Errorf("expected %d ID and %d name lookups, got %d and %d", tt.idCalls, tt.nameCalls, idCalls, nameCalls)
			}
		})
	}
}

// Test_tmWriteGuards checks that write operations reject missing configuration before reaching the API
func Test_tmWriteGuards(t *testing.T) {
	vcdClient := &VCDClient{}
	avi := &TmAviController{TmAviController: &types.TmAviController{ID: "urn:vcloud:aviController:1"}, vcdClient: vcdClient}
	vxlan := &TmDistributedVxlanConnection{TmDistributedVxlanConnection: &types.TmDistributedVxlanConnection{ID: "urn:vcloud:vxlan:1"}, vcdClient: vcdClient}
	infra := &TmInfraPolicy{TmInfraPolicy: &types.TmInfraPolicy{ID: "urn:vcloud:infraPolicy:1"}, vcdClient: vcdClient}
	namespace := &TmNamespace{TmNamespace: &types.TmNamespace{ID: "urn:vcloud:namespace:1"}, vcdClient: vcdClient}
	project := &TmProjectAssignment{TmProjectAssignment: &types.TmProjectAssignment{ID: "urn:vcloud:projectAssignment:1"}, vcdClient: vcdClient}

	checks := map[string]func() error{
		"CreateTmAviController":                        func() error { _, err := vcdClient.CreateTmAviController(nil); return err },
		"CreateTmAviControllerAsync":                   func() error { _, err := vcdClient.CreateTmAviControllerAsync(nil); return err },
		"CreateTmAviServiceEngineGroupAssignment":      func() error { _, err := vcdClient.CreateTmAviServiceEngineGroupAssignment(nil); return err },
		"CreateTmAviServiceEngineGroupAssignmentAsync": func() error { _, err := vcdClient.CreateTmAviServiceEngineGroupAssignmentAsync(nil); return err },
		"CreateTmDistributedVxlanConnection":           func() error { _, err := vcdClient.CreateTmDistributedVxlanConnection(nil); return err },
		"CreateTmDistributedVxlanConnectionAsync":      func() error { _, err := vcdClient.CreateTmDistributedVxlanConnectionAsync(nil); return err },
		"CreateTmInfraPolicy":                          func() error { _, err := vcdClient.CreateTmInfraPolicy(nil); return err },
		"CreateTmInfraPolicyAsync":                     func() error { _, err := vcdClient.CreateTmInfraPolicyAsync(nil); return err },
		"CreateTmNamespace":                            func() error { _, err := vcdClient.CreateTmNamespace(nil); return err },
		"CreateTmNamespaceAsync":                       func() error { _, err := vcdClient.CreateTmNamespaceAsync(nil); return err },
		"ImportTmNamespace":                            func() error { return vcdClient.ImportTmNamespace(nil) },
		"ImportTmNamespaceAsync":                       func() error { _, err := vcdClient.ImportTmNamespaceAsync(nil); return err },
		"CreateTmProjectAssignment":                    func() error { _, err := vcdClient.CreateTmProjectAssignment(nil); return err },
		"CreateTmProjectAssignmentWithoutId":           func() error { _, err := vcdClient.CreateTmProjectAssignment(&types.TmProjectAssignment{}); return err },
		"CreateTmProjectAssignmentAsync":               func() error { _, err := vcdClient.CreateTmProjectAssignmentAsync(nil); return err },
		"TmAviController.Update":                       func() error { _, err := avi.Update(nil); return err },
		"TmAviController.UpdateAsync":                  func() error { _, err := avi.UpdateAsync(nil); return err },
		"TmDistributedVxlanConnection.Update":          func() error { _, err := vxlan.Update(nil); return err },
		"TmDistributedVxlanConnection.UpdateAsync":     func() error { _, err := vxlan.UpdateAsync(nil); return err },
		"TmInfraPolicy.Update":                         func() error { _, err := infra.Update(nil); return err },
		"TmNamespace.Update":                           func() error { _, err := namespace.Update(nil); return err },
		"TmNamespace.UpdateAsync":                      func() error { _, err := namespace.UpdateAsync(nil); return err },
		"TmProjectAssignment.Update":                   func() error { _, err := project.Update(nil); return err },
		"TmProjectAssignment.UpdateAsync":              func() error { _, err := project.UpdateAsync(nil); return err },
		"TmProjectAssignment.UpdateMembers":            func() error { return project.UpdateMembers(nil) },
		"TmProjectAssignment.UpdateMembersAsync":       func() error { _, err := project.UpdateMembersAsync(nil); return err },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			if err := check(); err == nil || !strings.Contains(err.Error(), "config") {
				t.Errorf("expected a missing config error, got %v", err)
			}
		})
	}
}
