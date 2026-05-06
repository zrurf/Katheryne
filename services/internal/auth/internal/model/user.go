package model

import "time"

type User struct {
	UID          int64
	Name         string
	Phone        string
	Avatar       *string
	Status       string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	LastLogin    time.Time
	OpaqueRecord []byte
}

type UserConfig struct {
	UID             int64
	Language        string
	MsgNotification bool
	SoundEnabled    bool
	AutoTranslate   bool
	TranslateTarget *string
	ContentFilter   bool
	Enable2FA       bool
	TOTPSecret      *string
	TOTPBackupCodes []string
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

type LoginLog struct {
	UID       *int64
	User      string
	Success   bool
	Reason    *string
	IP        *string
	UserAgent *string
}

type BanRecord struct {
	ID        int64
	UID       int64
	Reason    string
	CreatedAt time.Time
	ExpiredAt *time.Time
}

type UserDevice struct {
	ID         int64
	UID        int64
	DeviceID   string
	DeviceName *string
	Platform   *string
	PushToken  *string
	LastActive *time.Time
	CreatedAt  time.Time
}

type LoginSession struct {
	User string
	MAC  []byte
}
