package middleware

import "testing"

type test struct {
	name                string
	requiredPermissions []string
	inputPermissions    PermissionSet
	errMessage          string
}

func TestHasAll(t *testing.T) {
	tests := []test{
		{
			name:                "PassingCase",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:read"}),
			errMessage:          "",
		},
		{
			name:                "PassingMultiple",
			requiredPermissions: []string{"recipe:read", "recipe:write"},
			inputPermissions:    NewPermissionSet([]string{"recipe:read", "recipe:write"}),
			errMessage:          "",
		},
		{
			name:                "FailingCase",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:write"}),
			errMessage:          "MISSING_PERMISSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasAll := NewHasAll(tc.requiredPermissions)

			err := hasAll.Validate(tc.inputPermissions)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}
		})
	}
}

func TestHasOne(t *testing.T) {
	tests := []test{
		{
			name:                "PassingCase",
			requiredPermissions: []string{"recipe:read", "recipe:write"},
			inputPermissions:    NewPermissionSet([]string{"recipe:read"}),
			errMessage:          "",
		},
		{
			name:                "FailingCase",
			requiredPermissions: []string{"recipe:read", "recipe:write"},
			inputPermissions:    NewPermissionSet([]string{"recipe:delete"}),
			errMessage:          "MISSING_PERMISSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasOne := NewHasOne(tc.requiredPermissions)

			err := hasOne.Validate(tc.inputPermissions)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}
		})
	}
}

func TestHas(t *testing.T) {
	tests := []test{
		{
			name:                "PassingCase",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:read"}),
			errMessage:          "",
		},
		{
			name:                "PassingCaseMultiple",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:read", "recipe:write"}),
			errMessage:          "",
		},
		{
			name:                "FailingCase",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:write"}),
			errMessage:          "MISSING_PERMISSION",
		},
		{
			name:                "FailingCaseMultiple",
			requiredPermissions: []string{"recipe:read"},
			inputPermissions:    NewPermissionSet([]string{"recipe:write", "recipe:delete"}),
			errMessage:          "MISSING_PERMISSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			has := NewHas(tc.requiredPermissions[0])

			err := has.Validate(tc.inputPermissions)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}
		})
	}
}
