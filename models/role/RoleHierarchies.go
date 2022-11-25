package model_role

import (
	"start/datasources"

	"github.com/pkg/errors"
)

type RoleHierarchy struct {
	Uid           int   `gorm:"primary_key;auto_increment" json:"uid"`
	ParentRoleUid int64 `gorm:"index;default:-1" json:"parent_role_uid"`
	RoleUid       int64 `gorm:"index;not null" json:"role_uid"`
}

// ======= CRUD ===========
func (item *RoleHierarchy) Create() error {
	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *RoleHierarchy) Update() error {
	db := datasources.GetDatabaseAuth()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

// Given a list of roles, retrieve all distinct sub role.
func GetAllSubRoleUids(roleId int) ([]int, error) {
	resultSet := make([]struct{ SubGroup int }, 0)
	sql := `
	WITH RECURSIVE hierarchies(role_uid) AS (
		SELECT    role_uid
		FROM      role_hierarchies
		WHERE     role_uid = ?
		UNION
		SELECT    g.role_uid
		FROM      hierarchies AS h
		JOIN      role_hierarchies AS g
		ON        h.role_uid = g.parent_role_uid
	)
	SELECT DISTINCT(role_uid) AS sub_group FROM hierarchies
	WHERE role_uid <> ? -- Ignore parent groups.`

	db := datasources.GetDatabaseAuth().Debug().Raw(sql, roleId, roleId).Scan(&resultSet)

	subRoleUids := make([]int, len(resultSet))
	for i, row := range resultSet {
		subRoleUids[i] = row.SubGroup
	}

	return subRoleUids, db.Error
}

func (item *RoleHierarchy) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *RoleHierarchy) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Uid <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
