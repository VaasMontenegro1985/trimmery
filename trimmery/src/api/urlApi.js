const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

async function parseResponse(response) {
  const data = await response.json().catch(() => null);

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error("Сессия истекла. Войдите заново.");
    }

    throw new Error(
      data?.error?.message || `Request failed with status ${response.status}`,
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

  if (customCode.trim()) {
    body.custom_code = customCode.trim();
  }

  if (isValidToken(token)) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}/shorten`, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });

  return parseResponse(response);
}

export async function registerUser(email, password) {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  return parseResponse(response);
}

export async function loginUser(email, password) {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  return parseResponse(response);
}

export async function getUserLinks(token) {
  const response = await fetch(`${API_BASE_URL}/api/links`, {
    method: "GET",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  return parseResponse(response);
}

export async function getLinkQrBlob(code, token, size = 256) {
  const response = await fetch(`${API_BASE_URL}/api/links/${code}/qr?size=${size}`, {
    method: "GET",
    headers: {
      ...(isValidToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error("Сессия истекла. Войдите заново.");
    }

    const data = await response.json().catch(() => null);

    throw new Error(
      data?.error?.message || `Request failed with status ${response.status}`,
    );
  }

  return response.blob();
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

  return parseResponse(response);
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

  if (response.status === 401) {
    throw new Error("Сессия истекла. Войдите заново.");
  }

  const data = await response.json().catch(() => null);

  throw new Error(
    data?.error?.message || `Request failed with status ${response.status}`,
  );
}
