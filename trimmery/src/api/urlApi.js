const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

export const SESSION_EXPIRED_MESSAGE = "Сессия истекла. Войдите заново.";

async function parseResponse(response, options = {}) {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    if (response.status === 401) {
      if (options.sessionProtected) {
        throw new Error(SESSION_EXPIRED_MESSAGE);
      }

      if (options.authEndpoint) {
        throw new Error(
          data?.error?.message ||
            options.authFallbackMessage ||
            `Запрос завершился ошибкой ${response.status}`,
        );
      }
    }

    throw new Error(
      data?.error?.message || `Запрос завершился ошибкой ${response.status}`,
    );
  }

  return data;
}

export function isValidToken(token) {
  return typeof token === "string" && token.trim() !== "";
}

export async function shortenUrl(url, customCode = "", token = null) {
  const body = { url };
  const headers = {
    "Content-Type": "application/json",
  };
  const hasAuthorization = isValidToken(token);

  if (customCode.trim()) {
    body.custom_code = customCode.trim();
  }

  if (hasAuthorization) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}/shorten`, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });

  return parseResponse(response, { sessionProtected: hasAuthorization });
}

export async function registerUser(email, password) {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  return parseResponse(response, { authEndpoint: true });
}

export async function loginUser(email, password) {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  return parseResponse(response, {
    authEndpoint: true,
    authFallbackMessage: "Неверная почта или пароль",
  });
}

export async function getUserLinks(token) {
  const response = await fetch(`${API_BASE_URL}/api/links`, {
    method: "GET",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  return parseResponse(response, { sessionProtected: true });
}

export async function getLinkQrBlob(code, token, size = 256) {
  const response = await fetch(`${API_BASE_URL}/api/links/${code}/qr?size=${size}`, {
    method: "GET",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  if (!response.ok) {
    await parseResponse(response, { sessionProtected: true });
  }

  return response.blob();
}

export async function getLinkStats(code, token, limit = 10) {
  const response = await fetch(`${API_BASE_URL}/api/links/${code}/stats?limit=${limit}`, {
    method: "GET",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  return parseResponse(response, { sessionProtected: true });
}

export async function updateLink(code, payload, token) {
  const response = await fetch(`${API_BASE_URL}/api/links/${code}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(payload),
  });

  return parseResponse(response, { sessionProtected: true });
}

export async function deleteLink(code, token) {
  const response = await fetch(`${API_BASE_URL}/api/links/${code}`, {
    method: "DELETE",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  if (response.status === 204) {
    return true;
  }

  if (!response.ok) {
    await parseResponse(response, { sessionProtected: true });
  }

  return true;
}
