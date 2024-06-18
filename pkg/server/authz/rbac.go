package authz

import (
	"github.com/njayp/jobber/pkg/pb"
)

// authorized users
const (
	alice = "alice@gmail.com"
	bob   = "bob@gmail.com"
)

// roles
// writer can Start and Stop
// reader can Status and Stream
var (
	reader = role{name: "reader", permissions: []string{pb.Jobber_Start_FullMethodName, pb.Jobber_Stop_FullMethodName}}
	writer = role{name: "writer", permissions: []string{pb.Jobber_Status_FullMethodName, pb.Jobber_Stream_FullMethodName}}
)

type role struct {
	name        string
	permissions []string
}

type roleBinding struct {
	userName string
	roleName string
}

// RBAC system
type RBAC struct {
	roles        map[string]role
	roleBindings []roleBinding
}

// Create a new RBAC system
func NewRBAC() *RBAC {
	return &RBAC{
		roles: map[string]role{
			reader.name: reader,
			writer.name: writer,
		},
		roleBindings: []roleBinding{
			// make alice a writer, reader
			{alice, reader.name},
			{alice, writer.name},
			// make bob a reader
			{bob, reader.name},
		},
	}
}

// get all rolebindings for user
func (r *RBAC) getUserRoles(userName string) []string {
	var roles []string
	for _, rb := range r.roleBindings {
		if rb.userName == userName {
			roles = append(roles, rb.roleName)
		}
	}
	// TODO rm duplicates
	return roles
}

// Check if a role has access to a specific rpc
func (r *RBAC) roleHasAccess(roleName, rpcTarget string) bool {
	role, exists := r.roles[roleName]
	if !exists {
		return false
	}

	for _, rpc := range role.permissions {
		if rpc == rpcTarget {
			return true
		}
	}
	return false
}

// Authorize user for a specific rpc
func (r *RBAC) Authorize(user, rpcTarget string) bool {
	for _, roleName := range r.getUserRoles(user) {
		if r.roleHasAccess(roleName, rpcTarget) {
			return true
		}
	}
	return false
}
