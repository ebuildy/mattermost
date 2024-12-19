// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"
)

func TestGetCPAField(t *testing.T) {
	os.Setenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES", "true")
	defer os.Unsetenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES")
	th := Setup(t).InitBasic()
	defer th.TearDown()

	cpaGroup, rErr := th.App.Srv().propertyService.RegisterPropertyGroup(model.CustomProfileAttributesPropertyGroupName)
	require.NoError(t, rErr)

	t.Run("should fail when getting a non-existent field", func(t *testing.T) {
		field, err := th.App.GetCPAField(model.NewId())
		require.NotNil(t, err)
		require.Equal(t, "app.custom_profile_attributes.get_property_field.app_error", err.Id)
		require.Empty(t, field)
	})

	t.Run("should fail when getting a field from a different group", func(t *testing.T) {
		field := &model.PropertyField{
			GroupID: model.NewId(),
			Name:    model.NewId(),
			Type:    model.PropertyFieldTypeText,
		}
		createdField, err := th.App.Srv().propertyService.CreatePropertyField(field)
		require.NoError(t, err)

		fetchedField, appErr := th.App.GetCPAField(createdField.ID)
		require.NotNil(t, appErr)
		require.Equal(t, "app.custom_profile_attributes.property_field_not_found.app_error", appErr.Id)
		require.Empty(t, fetchedField)
	})

	t.Run("should get an existing CPA field", func(t *testing.T) {
		field := &model.PropertyField{
			GroupID: cpaGroup.ID,
			Name:    "Test Field",
			Type:    model.PropertyFieldTypeText,
			Attrs:   map[string]any{"visibility": "hidden"},
		}

		createdField, err := th.App.CreateCPAField(field)
		require.Nil(t, err)
		require.NotEmpty(t, createdField.ID)

		fetchedField, err := th.App.GetCPAField(createdField.ID)
		require.Nil(t, err)
		require.Equal(t, createdField.ID, fetchedField.ID)
		require.Equal(t, "Test Field", fetchedField.Name)
		require.Equal(t, map[string]any{"visibility": "hidden"}, fetchedField.Attrs)
	})
}

func TestListCPAFields(t *testing.T) {
	os.Setenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES", "true")
	defer os.Unsetenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES")
	th := Setup(t).InitBasic()
	defer th.TearDown()

	cpaGroup, rErr := th.App.Srv().propertyService.RegisterPropertyGroup(model.CustomProfileAttributesPropertyGroupName)
	require.NoError(t, rErr)

	t.Run("should list the CPA property fields", func(t *testing.T) {
		field1 := &model.PropertyField{
			GroupID: cpaGroup.ID,
			Name:    "Field 1",
			Type:    model.PropertyFieldTypeText,
		}

		_, err := th.App.Srv().propertyService.CreatePropertyField(field1)
		require.NoError(t, err)

		field2 := &model.PropertyField{
			GroupID: model.NewId(),
			Name:    "Field 2",
			Type:    model.PropertyFieldTypeText,
		}
		_, err = th.App.Srv().propertyService.CreatePropertyField(field2)
		require.NoError(t, err)

		field3 := &model.PropertyField{
			GroupID: cpaGroup.ID,
			Name:    "Field 3",
			Type:    model.PropertyFieldTypeText,
		}
		_, err = th.App.Srv().propertyService.CreatePropertyField(field3)
		require.NoError(t, err)

		fields, err := th.App.ListCPAFields()
		require.Nil(t, err)
		require.Len(t, fields, 2)

		fieldNames := []string{}
		for _, field := range fields {
			fieldNames = append(fieldNames, field.Name)
		}
		require.ElementsMatch(t, []string{"Field 1", "Field 3"}, fieldNames)
	})
}

func TestCreateCPAField(t *testing.T) {
	os.Setenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES", "true")
	defer os.Unsetenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES")
	th := Setup(t).InitBasic()
	defer th.TearDown()

	cpaGroup, rErr := th.App.Srv().propertyService.RegisterPropertyGroup(model.CustomProfileAttributesPropertyGroupName)
	require.NoError(t, rErr)

	t.Run("should fail if the field is not valid", func(t *testing.T) {
		field := &model.PropertyField{Name: model.NewId()}

		createdField, err := th.App.CreateCPAField(field)
		require.NotNil(t, err)
		require.Empty(t, createdField)
	})

	t.Run("should not be able to create a property field for a different feature", func(t *testing.T) {
		field := &model.PropertyField{
			GroupID: model.NewId(),
			Name:    model.NewId(),
			Type:    model.PropertyFieldTypeText,
		}

		createdField, err := th.App.CreateCPAField(field)
		require.Nil(t, err)
		require.Equal(t, cpaGroup.ID, createdField.GroupID)
	})

	t.Run("should correctly create a CPA field", func(t *testing.T) {
		field := &model.PropertyField{
			GroupID: cpaGroup.ID,
			Name:    model.NewId(),
			Type:    model.PropertyFieldTypeText,
			Attrs:   map[string]any{"visibility": "hidden"},
		}

		createdField, err := th.App.CreateCPAField(field)
		require.Nil(t, err)
		require.NotZero(t, createdField.ID)
		require.Equal(t, cpaGroup.ID, createdField.GroupID)
		require.Equal(t, map[string]any{"visibility": "hidden"}, createdField.Attrs)

		fetchedField, gErr := th.App.Srv().propertyService.GetPropertyField(createdField.ID)
		require.NoError(t, gErr)
		require.Equal(t, field.Name, fetchedField.Name)
		require.NotZero(t, fetchedField.CreateAt)
		require.Equal(t, fetchedField.CreateAt, fetchedField.UpdateAt)
	})

	t.Run("should not be able to create CPA fields above the limit", func(t *testing.T) {
		// we create the rest of the fields required to reach the limit
		for i := 1; i < CustomProfileAttributesFieldLimit-1; i++ {
			field := &model.PropertyField{
				Name: model.NewId(),
				Type: model.PropertyFieldTypeText,
			}
			createdField, err := th.App.CreateCPAField(field)
			require.Nil(t, err)
			require.NotZero(t, createdField.ID)
		}

		// then, we create a last one that would exceed the limit
		field := &model.PropertyField{
			Name: model.NewId(),
			Type: model.PropertyFieldTypeText,
		}
		createdField, err := th.App.CreateCPAField(field)
		require.NotNil(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, err.StatusCode)
		require.Zero(t, createdField)
	})
}

func TestPatchCPAField(t *testing.T) {
	os.Setenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES", "true")
	defer os.Unsetenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES")
	th := Setup(t).InitBasic()
	defer th.TearDown()

	cpaGroup, rErr := th.App.Srv().propertyService.RegisterPropertyGroup(model.CustomProfileAttributesPropertyGroupName)
	require.NoError(t, rErr)

	newField := &model.PropertyField{
		GroupID: cpaGroup.ID,
		Name:    model.NewId(),
		Type:    model.PropertyFieldTypeText,
		Attrs:   map[string]any{"visibility": "hidden"},
	}
	field, err := th.App.CreateCPAField(newField)
	require.Nil(t, err)

	patch := &model.PropertyFieldPatch{
		Name:       model.NewPointer("Patched name"),
		Attrs:      model.NewPointer(map[string]any{"visibility": "default"}),
		TargetID:   model.NewPointer(model.NewId()),
		TargetType: model.NewPointer(model.NewId()),
	}

	t.Run("should fail if the field doesn't exist", func(t *testing.T) {
		updatedField, err := th.App.PatchCPAField(model.NewId(), patch)
		require.NotNil(t, err)
		require.Empty(t, updatedField)
	})

	t.Run("should not allow to patch a field outside of CPA", func(t *testing.T) {
		newField := &model.PropertyField{
			GroupID: model.NewId(),
			Name:    model.NewId(),
			Type:    model.PropertyFieldTypeText,
		}
		field, err := th.App.Srv().propertyService.CreatePropertyField(newField)
		require.NoError(t, err)

		updatedField, uErr := th.App.PatchCPAField(field.ID, patch)
		require.NotNil(t, uErr)
		require.Equal(t, "app.custom_profile_attributes.property_field_not_found.app_error", uErr.Id)
		require.Empty(t, updatedField)
	})

	t.Run("should correctly patch the CPA property field", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // ensure the UpdateAt is different than CreateAt

		updatedField, err := th.App.PatchCPAField(field.ID, patch)
		require.Nil(t, err)
		require.Equal(t, field.ID, updatedField.ID)
		require.Equal(t, "Patched name", updatedField.Name)
		require.Equal(t, "default", updatedField.Attrs["visibility"])
		require.Empty(t, updatedField.TargetID, "CPA should not allot to patch the field's target ID")
		require.Empty(t, updatedField.TargetType, "CPA should not allot to patch the field's target type")
		require.Greater(t, updatedField.UpdateAt, field.UpdateAt)
	})
}

func TestDeleteCPAField(t *testing.T) {
	os.Setenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES", "true")
	defer os.Unsetenv("MM_FEATUREFLAGS_CUSTOMPROFILEATTRIBUTES")
	th := Setup(t).InitBasic()
	defer th.TearDown()

	cpaGroup, rErr := th.App.Srv().propertyService.RegisterPropertyGroup(model.CustomProfileAttributesPropertyGroupName)
	require.NoError(t, rErr)

	newField := &model.PropertyField{
		GroupID: cpaGroup.ID,
		Name:    model.NewId(),
		Type:    model.PropertyFieldTypeText,
	}
	field, err := th.App.CreateCPAField(newField)
	require.Nil(t, err)

	for i := 0; i < 3; i++ {
		newValue := &model.PropertyValue{
			TargetID:   model.NewId(),
			TargetType: "user",
			GroupID:    cpaGroup.ID,
			FieldID:    field.ID,
			Value:      fmt.Sprintf("Value %d", i),
		}
		value, err := th.App.Srv().propertyService.CreatePropertyValue(newValue)
		require.NoError(t, err)
		require.NotZero(t, value.ID)
	}

	t.Run("should fail if the field doesn't exist", func(t *testing.T) {
		err := th.App.DeleteCPAField(model.NewId())
		require.NotNil(t, err)
		require.Equal(t, "app.custom_profile_attributes.get_property_field.app_error", err.Id)
	})

	t.Run("should not allow to delete a field outside of CPA", func(t *testing.T) {
		newField := &model.PropertyField{
			GroupID: model.NewId(),
			Name:    model.NewId(),
			Type:    model.PropertyFieldTypeText,
		}
		field, err := th.App.Srv().propertyService.CreatePropertyField(newField)
		require.NoError(t, err)

		dErr := th.App.DeleteCPAField(field.ID)
		require.NotNil(t, dErr)
		require.Equal(t, "app.custom_profile_attributes.property_field_not_found.app_error", dErr.Id)
	})

	t.Run("should correctly delete the field", func(t *testing.T) {
		// check that we have the associated values to the field prior deletion
		opts := model.PropertyValueSearchOpts{PerPage: 10, FieldID: field.ID}
		values, err := th.App.Srv().propertyService.SearchPropertyValues(opts)
		require.NoError(t, err)
		require.Len(t, values, 3)

		// delete the field
		require.Nil(t, th.App.DeleteCPAField(field.ID))

		// check that it is marked as deleted
		fetchedField, err := th.App.Srv().propertyService.GetPropertyField(field.ID)
		require.Nil(t, err)
		require.NotZero(t, fetchedField.DeleteAt)

		// ensure that the associated fields have been marked as deleted too
		values, err = th.App.Srv().propertyService.SearchPropertyValues(opts)
		require.NoError(t, err)
		require.Len(t, values, 0)

		opts.IncludeDeleted = true
		values, err = th.App.Srv().propertyService.SearchPropertyValues(opts)
		require.NoError(t, err)
		require.Len(t, values, 3)
		for _, value := range values {
			require.NotZero(t, value.DeleteAt)
		}
	})
}
