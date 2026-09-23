package service

import (
	"latihan-repository/app/model"
	"latihan-repository/helper"
)

// mengizinkan pemilik data, atau role yang punya permission :any.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}

	return perms.Can(current.Role, anyPermission)
}
