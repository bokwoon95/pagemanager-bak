package tables

import (
	"context"

	"github.com/bokwoon95/pagemanager/sq"
)

type TenantIDKey struct{}

type PM_SUPERADMIN struct {
	sq.TableInfo
	ID                        sq.NumberField `sq:"type=INTEGER misc=PRIMARY_KEY"`
	PASSWORD_HASH             sq.StringField
	ENCRYPTION_KEY_PARAMETERS sq.StringField
}

func NEW_SUPERADMIN(ctx context.Context, alias string) PM_SUPERADMIN {
	tbl := PM_SUPERADMIN{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_superadmin"
	} else {
		tbl.TableInfo.Name = "pm_superadmin"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_ENCRYPTION_KEYS struct {
	sq.TableInfo
	ID             sq.NumberField `sq:"type=INTEGER misc=NOT_NULL,UNIQUE"`
	KEY_CIPHERTEXT sq.StringField
	CREATED_AT     sq.TimeField
}

func NEW_ENCRYPTION_KEYS(ctx context.Context, alias string) PM_ENCRYPTION_KEYS {
	tbl := PM_ENCRYPTION_KEYS{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_encryption_keys"
	} else {
		tbl.TableInfo.Name = "pm_encryption_keys"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_MAC_KEYS struct {
	sq.TableInfo
	ID             sq.NumberField `sq:"type=INTEGER misc=NOT_NULL,UNIQUE"`
	KEY_CIPHERTEXT sq.StringField
	CREATED_AT     sq.TimeField
}

func NEW_MAC_KEYS(ctx context.Context, alias string) PM_MAC_KEYS {
	tbl := PM_MAC_KEYS{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_mac_keys"
	} else {
		tbl.TableInfo.Name = "pm_mac_keys"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_PAGES struct {
	sq.TableInfo
	URL sq.StringField `sq:"type=TEXT misc=NOT_NULL,PRIMARY_KEY"`
	// 404 Not Found
	DISABLED sq.BooleanField
	// 301 Moved Permanently
	REDIRECT_URL sq.StringField
	// plugins
	PLUGIN       sq.StringField
	HANDLER_NAME sq.StringField
	HANDLER_URL  sq.StringField
	// content body
	CONTENT sq.StringField
	// templates
	THEME_PATH sq.StringField
	TEMPLATE   sq.StringField
}

func NEW_PAGES(ctx context.Context, alias string) PM_PAGES {
	tbl := PM_PAGES{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_pages"
	} else {
		tbl.TableInfo.Name = "pm_pages"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_PAGEDATA struct {
	sq.TableInfo
	LOCALE_CODE sq.StringField `sq:"misc=NOT_NULL"`
	DATA_ID     sq.StringField `sq:"misc=NOT_NULL"`
	KEY         sq.StringField `sq:"misc=NOT_NULL"`
	VALUE       sq.JSONField   `sq:"misc=NOT_NULL"`
	ARRAY_INDEX sq.NumberField `sq:""`
}

func NEW_PAGEDATA(ctx context.Context, alias string) PM_PAGEDATA {
	tbl := PM_PAGEDATA{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_pagedata"
	} else {
		tbl.TableInfo.Name = "pm_pagedata"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_USERS struct {
	sq.TableInfo
	USER_ID        sq.NumberField `sq:"type=INTEGER misc=PRIMARY_KEY"`
	PUBLIC_USER_ID sq.StringField `sq:"type=TEXT misc=NOT_NULL,UNIQUE"`
	USERNAME       sq.StringField
	PASSWORD_HASH  sq.StringField
	AUTHZ_DATA     sq.JSONField
	AUTHZ_GROUPS   sq.JSONField
	USER_DATA      sq.JSONField
}

func NEW_USERS(ctx context.Context, alias string) PM_USERS {
	tbl := PM_USERS{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_users"
	} else {
		tbl.TableInfo.Name = "pm_users"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_AUTHZ_GROUPS struct {
	sq.TableInfo
	NAME       sq.StringField `sq:"type=TEXT misc=NOT_NULL,PRIMARY_KEY"`
	AUTHZ_DATA sq.JSONField
}

func NEW_AUTHZ_GROUPS(ctx context.Context, alias string) PM_AUTHZ_GROUPS {
	tbl := PM_AUTHZ_GROUPS{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_authz_groups"
	} else {
		tbl.TableInfo.Name = "pm_authz_groups"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_SESSIONS struct {
	sq.TableInfo
	SESSION_HASH sq.StringField `sq:"type=TEXT misc=NOT_NULL,PRIMARY_KEY"`
	USER_ID      sq.NumberField `sq:"type=INTEGER misc=NOT_NULL"`
	CREATED_AT   sq.TimeField
	SESSION_DATA sq.JSONField
}

func NEW_SESSIONS(ctx context.Context, alias string) PM_SESSIONS {
	tbl := PM_SESSIONS{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_sessions"
	} else {
		tbl.TableInfo.Name = "pm_sessions"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}

type PM_LOCALES struct {
	sq.TableInfo
	LOCALE_CODE sq.StringField `sq:"type=TEXT misc=PRIMARY_KEY"`
	DESCRIPTION sq.StringField
}

func NEW_LOCALES(ctx context.Context, alias string) PM_LOCALES {
	tbl := PM_LOCALES{TableInfo: sq.TableInfo{Alias: alias}}
	if tenantID, ok := ctx.Value(TenantIDKey{}).(string); ok && tenantID != "" {
		tbl.TableInfo.Name = "pm_" + tenantID + "_locales"
	} else {
		tbl.TableInfo.Name = "pm_locales"
	}
	_ = sq.ReflectTable(&tbl)
	return tbl
}
