package permission

// Permission represents a permission string
type Permission string

// Game permissions
const (
	GameReleaseCreate Permission = "game:release:create"
	GameReleaseUpdate Permission = "game:release:update"
	GameReleaseDelete Permission = "game:release:delete"
	GameView          Permission = "game:view"
	GameManage        Permission = "game:manage"
)

// Comment permissions
const (
	CommentCreate    Permission = "comment:create"
	CommentDeleteOwn Permission = "comment:delete:own"
	CommentDeleteAny Permission = "comment:delete:any"
	CommentUpdate    Permission = "comment:update"
)

// OST permissions
const (
	OSTView         Permission = "ost:view"
	OSTDownloadHiFi Permission = "ost:download:hifi"
	OSTManage       Permission = "ost:manage"
)

// Dashboard permissions
const (
	DashboardView Permission = "dashboard:view"
)

// User permissions
const (
	UserView   Permission = "user:view"
	UserUpdate Permission = "user:update"
	UserDelete Permission = "user:delete"
	UserManage Permission = "user:manage"
)
