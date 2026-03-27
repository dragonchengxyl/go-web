export interface AccessTokenClaims {
  role?: string;
  permissions?: string[];
  force_password_reset?: boolean;
}

const ADMIN_CONSOLE_ROLES = new Set(["admin", "super_admin"]);
const ADMIN_CONSOLE_PERMISSIONS = new Set(["dashboard:view"]);

function decodeBase64URL(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padding = normalized.length % 4;
  const padded =
    padding === 0 ? normalized : `${normalized}${"=".repeat(4 - padding)}`;
  return atob(padded);
}

export function parseAccessTokenClaims(
  token?: string | null,
): AccessTokenClaims | null {
  if (!token || typeof atob !== "function") {
    return null;
  }

  const parts = token.split(".");
  if (parts.length < 2) {
    return null;
  }

  try {
    const payload = JSON.parse(decodeBase64URL(parts[1])) as {
      role?: unknown;
      permissions?: unknown;
      force_password_reset?: unknown;
    };

    return {
      role: typeof payload.role === "string" ? payload.role : undefined,
      permissions: Array.isArray(payload.permissions)
        ? payload.permissions.filter(
            (item): item is string => typeof item === "string",
          )
        : [],
      force_password_reset: Boolean(payload.force_password_reset),
    };
  } catch {
    return null;
  }
}

export function canAccessAdminConsole(input?: {
  role?: string | null;
  permissions?: string[] | null;
}) {
  const role = input?.role?.trim() ?? "";
  if (ADMIN_CONSOLE_ROLES.has(role)) {
    return true;
  }

  return (input?.permissions ?? []).some((perm) =>
    ADMIN_CONSOLE_PERMISSIONS.has(perm),
  );
}

export function shouldForceAdminPasswordReset(
  claims: AccessTokenClaims | null,
) {
  return claims?.role === "admin" && Boolean(claims.force_password_reset);
}
