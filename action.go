package tfplan_validator

import (
	tfjson "github.com/hashicorp/terraform-json"
)

// Action is an enum that corresponds to a valid tfjson plan array e.g. ["delete", "create"] is ActionDeleteCreate
type Action string

const (
	ActionInvalid             Action = "invalid"
	ActionNoOp                Action = "no-op"
	ActionRead                Action = "read"
	ActionCreate              Action = "create"
	ActionUpdate              Action = "update"
	ActionDelete              Action = "delete"
	ActionDestroyBeforeCreate Action = "delete-create"
	ActionCreateBeforeDestroy Action = "create-delete"
)

var pretty = map[Action]string{
	ActionCreate:              "created",
	ActionUpdate:              "updated",
	ActionDelete:              "deleted",
	ActionDestroyBeforeCreate: "replaced (deleted then re-created)",
	ActionCreateBeforeDestroy: "replaced (re-created before deletion)",
}

// CompatiblePairs are actions that may be equivalent between different state files
var compatiblePairs = map[[2]Action]bool{
	{ActionCreate, ActionUpdate}:              true,
	{ActionCreate, ActionDestroyBeforeCreate}: true,
	{ActionCreate, ActionCreateBeforeDestroy}: true,
	{ActionUpdate, ActionDestroyBeforeCreate}: true,
	{ActionUpdate, ActionCreateBeforeDestroy}: true,
}

// AreCompatible returns true if the actions may be equivalent between different state files
func AreCompatible(left Action, right Action) bool {
	return left == right ||
		compatiblePairs[[2]Action{left, right}] ||
		compatiblePairs[[2]Action{right, left}]
}

// Pretty printable string
func (a Action) Pretty() string {
	return pretty[a]
}

// ConvertAction from the tfjson form to one we can more easily work with
func ConvertAction(actions *tfjson.Actions) Action {
	if actions.NoOp() {
		return ActionNoOp
	} else if actions.Read() {
		return ActionRead
	} else if actions.Create() {
		return ActionCreate
	} else if actions.Update() {
		return ActionUpdate
	} else if actions.Delete() {
		return ActionDelete
	} else if actions.DestroyBeforeCreate() {
		return ActionDestroyBeforeCreate
	} else if actions.CreateBeforeDestroy() {
		return ActionCreateBeforeDestroy
	}
	return ActionInvalid
}
