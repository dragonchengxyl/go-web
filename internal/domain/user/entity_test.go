package user

import "testing"

func TestNextForcePasswordReset(t *testing.T) {
	tests := []struct {
		name     string
		current  bool
		previous Role
		next     Role
		want     bool
	}{
		{
			name:     "member promoted to admin is forced",
			current:  false,
			previous: RoleMember,
			next:     RoleAdmin,
			want:     true,
		},
		{
			name:     "admin keeps existing force flag",
			current:  true,
			previous: RoleAdmin,
			next:     RoleAdmin,
			want:     true,
		},
		{
			name:     "admin cleared after changing to super admin",
			current:  true,
			previous: RoleAdmin,
			next:     RoleSuperAdmin,
			want:     false,
		},
		{
			name:     "member role never forces reset",
			current:  true,
			previous: RoleAdmin,
			next:     RoleMember,
			want:     false,
		},
		{
			name:     "super admin downgraded to admin is forced",
			current:  false,
			previous: RoleSuperAdmin,
			next:     RoleAdmin,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextForcePasswordReset(tt.current, tt.previous, tt.next); got != tt.want {
				t.Fatalf("NextForcePasswordReset(%v, %q, %q) = %v, want %v", tt.current, tt.previous, tt.next, got, tt.want)
			}
		})
	}
}
