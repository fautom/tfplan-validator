package tfplan_validator

import (
	"path"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
)

func planPath(typ string) string {
	return path.Join("test", "fixtures", typ, "plan.json")
}

func filterPath(typ string) string {
	return path.Join("test", "fixtures", typ, "filter.json")
}

func otherPath(name string) string {
	return path.Join("test", "fixtures", "itest", name)
}

func readPlansP(paths []string) []*tfjson.Plan {
	if plans, err := ReadPlans(paths); err != nil {
		panic(err)
	} else {
		return plans
	}
}

func TestNewFilterFromPlans(t *testing.T) {
	cases := []struct {
		name     string
		in       []*tfjson.Plan
		expected *PlanFilter
		errStr   string
	}{
		{
			name: "empty",
			in:   []*tfjson.Plan{{}},
			expected: &PlanFilter{
				FormatVersion:  CurrentFormatVersion,
				AllowedActions: map[Address][]Action{},
			},
		},
		{
			name: "create",
			in:   readPlansP([]string{planPath("create")}),
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"local_file.foo": {ActionCreate},
				},
			},
		},
		{
			name: "create-delete",
			in:   readPlansP([]string{planPath("create-delete")}),
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"local_file.foo": {ActionCreateBeforeDestroy},
				},
			},
		},
		{
			name: "delete",
			in:   readPlansP([]string{planPath("delete")}),
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"local_file.foo": {ActionDelete},
				},
			},
		},
		{
			name: "delete-create",
			in:   readPlansP([]string{planPath("delete-create")}),
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"local_file.foo": {ActionDestroyBeforeCreate},
				},
			},
		},
		{
			name: "update",
			in:   readPlansP([]string{planPath("update")}),
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"google_project_iam_policy.project": {ActionUpdate},
				},
			},
		},
		{
			name: "missing plan",
			in: []*tfjson.Plan{
				{
					ResourceChanges: []*tfjson.ResourceChange{
						{
							Mode:    tfjson.ManagedResourceMode,
							Address: "a.b.c",
							Change: &tfjson.Change{
								Actions: tfjson.Actions{"invalid"},
							},
						},
					},
				},
			},
			expected: nil,
			errStr:   "unrecognized action in plan: [invalid]",
		},
		{
			name: "ignore data",
			in: []*tfjson.Plan{
				{
					ResourceChanges: []*tfjson.ResourceChange{
						{
							Mode:    tfjson.DataResourceMode,
							Address: "a.b.c",
						},
					},
				},
			},
			expected: &PlanFilter{
				FormatVersion:  CurrentFormatVersion,
				AllowedActions: map[Address][]Action{},
			},
		},
		{
			name: "create and create-delete are compatible",
			in: []*tfjson.Plan{
				{
					ResourceChanges: []*tfjson.ResourceChange{
						{
							Mode:    tfjson.ManagedResourceMode,
							Address: "a.b.c",
							Change: &tfjson.Change{
								Actions: tfjson.Actions{tfjson.ActionCreate},
							},
						},
					},
				},
				{
					ResourceChanges: []*tfjson.ResourceChange{
						{
							Mode:    tfjson.ManagedResourceMode,
							Address: "a.b.c",
							Change: &tfjson.Change{
								Actions: tfjson.Actions{tfjson.ActionCreate, tfjson.ActionDelete},
							},
						},
					},
				},
			},
			expected: &PlanFilter{
				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"a.b.c": {ActionCreate, ActionCreateBeforeDestroy},
				},
			},
		},
		{
			name:   "duplicate change in plan",
			in:     readPlansP([]string{otherPath("plan-duplicate-address.json")}),
			errStr: "duplicate address in plan: local_file.foo [create]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewFilterFromPlans(tc.in)
			errStr := makeErrStr(err)
			if !reflect.DeepEqual(tc.expected, actual) || tc.errStr != errStr {
				t.Fatalf("expected:\n\n%s\ngot:\n\n%s\n\nexpected err:%s\n\ngot err: %s\n", spew.Sdump(tc.expected), spew.Sdump(actual), tc.errStr, errStr)
			}
		})
	}
}

func TestReadPlanFilters(t *testing.T) {
	cases := []struct {
		name     string
		in       []string
		expected []*PlanFilter
		errStr   string
	}{
		{
			name: "load two files",
			in:   []string{filterPath("create"), filterPath("update")},
			expected: []*PlanFilter{
				{
					FormatVersion: CurrentFormatVersion,
					AllowedActions: map[Address][]Action{
						"local_file.foo": {ActionCreate},
					},
				},
				{
					FormatVersion: CurrentFormatVersion,
					AllowedActions: map[Address][]Action{
						"google_project_iam_policy.project": {ActionUpdate},
					},
				},
			},
		},
		{
			name:   "one file is missing",
			in:     []string{filterPath("create"), filterPath("missing")},
			errStr: "open " + filterPath("missing") + ": no such file or directory",
		},
		{
			name:   "one file is missing",
			in:     []string{filterPath("create"), otherPath("unparseable.json")},
			errStr: otherPath("unparseable.json") + ": unexpected end of JSON input",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ReadPlanFilters(tc.in)
			errStr := makeErrStr(err)
			if !reflect.DeepEqual(tc.expected, actual) || tc.errStr != errStr {
				t.Fatalf("expected:\n\n%s\ngot:\n\n%s\n\nexpected err:%s\n\ngot err: %s\n", spew.Sdump(tc.expected), spew.Sdump(actual), tc.errStr, errStr)
			}
		})
	}
}

// func (f *PlanFilter) HasAction(address Address, action Action) bool {
// 	allowed := f.AllowedActions[address]

// 	if allowed == nil {
// 		return false
// 	}

// 	for _, has := range allowed {
// 		if action == has {
// 			return true
// 		}
// 	}

// 	return false
// }

func TestFilterHasAction(t *testing.T) {
	cases := []struct {
		name     string
		self     *PlanFilter
		address  Address
		action   Action
		expected bool
	}{
		{
			name: "doesn't have address",
			self: &PlanFilter{

				FormatVersion:  CurrentFormatVersion,
				AllowedActions: map[Address][]Action{},
			},
			address:  "a.b.c",
			action:   ActionCreate,
			expected: false,
		},
		{
			name: "has address, doesn't have action",
			self: &PlanFilter{

				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"a.b.c": {ActionUpdate, ActionDelete},
				},
			},
			address:  "a.b.c",
			action:   ActionCreate,
			expected: false,
		},
		{
			name: "has action",
			self: &PlanFilter{

				FormatVersion: CurrentFormatVersion,
				AllowedActions: map[Address][]Action{
					"a.b.c": {ActionUpdate, ActionCreate},
				},
			},
			address:  "a.b.c",
			action:   ActionCreate,
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.self.HasAction(tc.address, tc.action)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Fatalf("expected:\n\n%s\ngot:\n\n%s\n\n", spew.Sdump(tc.expected), spew.Sdump(actual))
			}
		})
	}
}
