import { useEffect, useRef, useState } from "react";
import "./App.css";
import QRCode from "qrcode";
import {
  deleteLink,
  getLinkQrBlob,
  getUserLinks,
  isValidToken,
  loginUser,
  registerUser,
  shortenUrl,
  updateLink,
} from "./api/urlApi";

function Logo({ onClick }) {
  return (
    <button className="logo logoButton" onClick={onClick} type="button">
      <svg
        className="logoSvg"
        viewBox="0 0 48 48"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M36 26V38C36 39.0609 35.5786 40.0783 34.8284 40.8284C34.0783 41.5786 33.0609 42 32 42H10C8.93913 42 7.92172 41.5786 7.17157 40.8284C6.42143 40.0783 6 39.0609 6 38V16C6 14.9391 6.42143 13.9217 7.17157 13.1716C7.92172 12.4214 8.93913 12 10 12H22M30 6H42M42 6V18M42 6L20 28"
          stroke="#1E1E1E"
          strokeWidth="4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>

      <span className="logoText">TRIMMERY</span>
    </button>
  );
}

function formatDate(value) {
  if (!value) return "";

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(value));
}

function App() {
  const [screen, setScreen] = useState("home");
  const [authMode, setAuthMode] = useState("register");
  const [authEmail, setAuthEmail] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [username, setUsername] = useState("");
  const [authLoading, setAuthLoading] = useState(false);
  const [authError, setAuthError] = useState("");
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [expiresAt, setExpiresAt] = useState(null);
  const [url, setUrl] = useState("");
  const [customCode, setCustomCode] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [qrCode, setQrCode] = useState("");
  const [qrError, setQrError] = useState("");
  const [qrLoading, setQrLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [dashboardLinks, setDashboardLinks] = useState([]);
  const [dashboardLoading, setDashboardLoading] = useState(false);
  const [dashboardError, setDashboardError] = useState("");
  const [dashboardQr, setDashboardQr] = useState(null);
  const [dashboardQrLoadingCode, setDashboardQrLoadingCode] = useState("");
  const [dashboardQrError, setDashboardQrError] = useState(null);
  const [editingCode, setEditingCode] = useState("");
  const [editOriginalUrl, setEditOriginalUrl] = useState("");
  const [editAlias, setEditAlias] = useState("");
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");
  const [deleteLoadingCode, setDeleteLoadingCode] = useState("");
  const [dashboardMessage, setDashboardMessage] = useState("");
  const qrObjectUrlRef = useRef(null);
  const dashboardQrObjectUrlRef = useRef(null);

  const isTokenExpired = (value) => {
    if (!value) return true;

    const expiresTime = new Date(value).getTime();

    return Number.isNaN(expiresTime) || expiresTime <= Date.now();
  };

  const replaceQrCode = (nextQrCode) => {
    if (qrObjectUrlRef.current) {
      URL.revokeObjectURL(qrObjectUrlRef.current);
      qrObjectUrlRef.current = null;
    }

    if (nextQrCode?.startsWith("blob:")) {
      qrObjectUrlRef.current = nextQrCode;
    }

    setQrCode(nextQrCode);
  };

  const replaceDashboardQr = (nextDashboardQr) => {
    if (dashboardQrObjectUrlRef.current) {
      URL.revokeObjectURL(dashboardQrObjectUrlRef.current);
      dashboardQrObjectUrlRef.current = null;
    }

    if (nextDashboardQr?.dataUrl?.startsWith("blob:")) {
      dashboardQrObjectUrlRef.current = nextDashboardQr.dataUrl;
    }

    setDashboardQr(nextDashboardQr);
  };

  const clearSession = (message = "") => {
    setUser(null);
    setToken(null);
    setExpiresAt(null);
    setDashboardLinks([]);
    setEditingCode("");
    setEditError("");
    setDashboardMessage("");
    replaceDashboardQr(null);
    replaceQrCode("");
    localStorage.removeItem("trimmery_token");
    localStorage.removeItem("trimmery_user");
    localStorage.removeItem("trimmery_expires_at");

    if (message) {
      setError(message);
      setAuthError(message);
      setDashboardError(message);
    }
  };

  const hasActiveSession = () => {
    if (!isValidToken(token) || !user || isTokenExpired(expiresAt)) {
      clearSession("Сессия истекла. Войдите заново.");
      return false;
    }

    return true;
  };

  useEffect(() => {
    const savedToken = localStorage.getItem("trimmery_token");
    const savedUser = localStorage.getItem("trimmery_user");
    const savedExpiresAt = localStorage.getItem("trimmery_expires_at");

    if (!isValidToken(savedToken) || !savedUser || !savedExpiresAt) {
      clearSession();
      return;
    }

    if (isTokenExpired(savedExpiresAt)) {
      clearSession();
      return;
    }

    try {
      setUser(JSON.parse(savedUser));
      setToken(savedToken);
      setExpiresAt(savedExpiresAt);
    } catch {
      clearSession();
    }
  }, []);

  useEffect(() => {
    return () => {
      if (qrObjectUrlRef.current) {
        URL.revokeObjectURL(qrObjectUrlRef.current);
      }

      if (dashboardQrObjectUrlRef.current) {
        URL.revokeObjectURL(dashboardQrObjectUrlRef.current);
      }
    };
  }, []);

  const handleShorten = async () => {
    const inputValue = url.trim();
    const customCodeValue = customCode.trim();

    if (!inputValue || loading) return;

    setLoading(true);
    setError("");
    setQrError("");
    setQrLoading(false);

    try {
      const activeToken = isValidToken(token) && !isTokenExpired(expiresAt)
        ? token
        : null;

      if (isValidToken(token) && !activeToken) {
        clearSession("Сессия истекла. Войдите заново.");
        return;
      }

      const data = await shortenUrl(inputValue, customCodeValue, activeToken);

      setShortUrl(data.short_url);
      replaceQrCode("");

      if (activeToken) {
        setQrLoading(true);

        try {
          const qrBlob = await getLinkQrBlob(data.code, activeToken, 256);
          replaceQrCode(URL.createObjectURL(qrBlob));
        } catch (err) {
          if (err.message === "Сессия истекла. Войдите заново.") {
            clearSession(err.message);
          }

          setQrError("Не удалось загрузить QR-код");
        } finally {
          setQrLoading(false);
        }
      } else {
        // генерим QR в base64 PNG для анонимных ссылок
        const qr = await QRCode.toDataURL(data.short_url);
        replaceQrCode(qr);
      }
    } catch (err) {
      setError(err.message);
      setShortUrl("");
      replaceQrCode("");
      setQrError("");
    } finally {
      setLoading(false);
    }
  };

  const openLogin = () => {
    setAuthMode("login");
    setAuthError("");
    setScreen("auth");
  };

  const openHome = () => {
    setScreen("home");
  };

  const loadDashboardLinks = async (currentToken) => {
    setDashboardLoading(true);
    setDashboardError("");

    try {
      const data = await getUserLinks(currentToken);
      setDashboardLinks(data.links || []);
    } catch (err) {
      if (err.message === "Сессия истекла. Войдите заново.") {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

      setDashboardError(err.message);
      setDashboardLinks([]);
    } finally {
      setDashboardLoading(false);
    }
  };

  const openDashboard = async () => {
    if (!hasActiveSession()) {
      setAuthMode("login");
      setScreen("auth");
      return;
    }

    replaceDashboardQr(null);
    setDashboardQrError(null);
    setDashboardMessage("");
    setScreen("dashboard");
    await loadDashboardLinks(token);
  };

  const saveSession = (data) => {
    setToken(data.access_token);
    setUser(data.user);
    setExpiresAt(data.expires_at);
    localStorage.setItem("trimmery_token", data.access_token);
    localStorage.setItem("trimmery_user", JSON.stringify(data.user));
    localStorage.setItem("trimmery_expires_at", data.expires_at);
  };

  const handleAuthSubmit = async (event) => {
    event.preventDefault();

    if (authLoading) return;

    setAuthLoading(true);
    setAuthError("");

    try {
      const email = authEmail.trim();
      const password = authPassword;
      const data =
        authMode === "register"
          ? await registerUser(email, password)
          : await loginUser(email, password);

      saveSession(data);
      setAuthPassword("");
      setScreen("home");
    } catch (err) {
      setAuthError(err.message);
    } finally {
      setAuthLoading(false);
    }
  };

  const handleLogout = () => {
    clearSession();
    setScreen("home");
  };

  const switchAuthMode = (mode) => {
    setAuthMode(mode);
    setAuthError("");
  };

  const toggleDashboardQr = async (link) => {
    if (!hasActiveSession()) {
      setAuthMode("login");
      setScreen("auth");
      return;
    }

    if (dashboardQr?.code === link.code) {
      replaceDashboardQr(null);
      return;
    }

    setDashboardQrError(null);
    setDashboardQrLoadingCode(link.code);

    try {
      const qrBlob = await getLinkQrBlob(link.code, token, 256);
      replaceDashboardQr({
        code: link.code,
        dataUrl: URL.createObjectURL(qrBlob),
      });
    } catch (err) {
      if (err.message === "Сессия истекла. Войдите заново.") {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

      setDashboardQrError({
        code: link.code,
        message: "Не удалось загрузить QR-код",
      });
    } finally {
      setDashboardQrLoadingCode("");
    }
  };

  const startEditing = (link) => {
    setEditingCode(link.code);
    setEditOriginalUrl(link.original_url);
    setEditAlias(link.code);
    setEditError("");
    setDashboardMessage("");
  };

  const cancelEditing = () => {
    setEditingCode("");
    setEditOriginalUrl("");
    setEditAlias("");
    setEditError("");
  };

  const handleSaveEdit = async (link) => {
    if (!hasActiveSession()) {
      setAuthMode("login");
      setScreen("auth");
      return;
    }

    const payload = {};
    const nextOriginalUrl = editOriginalUrl.trim();
    const nextAlias = editAlias.trim();

    if (nextOriginalUrl !== link.original_url) {
      payload.original_url = nextOriginalUrl;
    }

    if (nextAlias !== link.code) {
      payload.code = nextAlias;
    }

    if (Object.keys(payload).length === 0) {
      setEditError("Нет изменений для сохранения");
      return;
    }

    setEditLoading(true);
    setEditError("");
    setDashboardMessage("");

    try {
      await updateLink(editingCode, payload, token);
      cancelEditing();
      replaceDashboardQr(null);
      await loadDashboardLinks(token);
      setDashboardMessage("Ссылка обновлена");
    } catch (err) {
      if (err.message === "Сессия истекла. Войдите заново.") {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

      setEditError(err.message);
    } finally {
      setEditLoading(false);
    }
  };

  const handleDelete = async (link) => {
    if (!hasActiveSession()) {
      setAuthMode("login");
      setScreen("auth");
      return;
    }

    if (!confirm("Удалить ссылку?")) {
      return;
    }

    setDeleteLoadingCode(link.code);
    setDashboardMessage("");
    setDashboardError("");

    try {
      await deleteLink(link.code, token);

      if (dashboardQr?.code === link.code) {
        replaceDashboardQr(null);
      }

      if (editingCode === link.code) {
        cancelEditing();
      }

      await loadDashboardLinks(token);
      setDashboardMessage("Ссылка удалена");
    } catch (err) {
      if (err.message === "Сессия истекла. Войдите заново.") {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

      setDashboardError(err.message);
    } finally {
      setDeleteLoadingCode("");
    }
  };

  const totalClicks = dashboardLinks.reduce(
    (sum, link) => sum + (link.clicks_count || 0),
    0,
  );
  const maxClicks = dashboardLinks.length
    ? Math.max(...dashboardLinks.map((link) => link.clicks_count || 0))
    : 0;

  if (screen === "dashboard") {
    return (
      <div className="page dashboardPage">
        <header className="header dashboardHeader">
          <Logo onClick={openHome} />

          <button className="outlineButton" onClick={openHome} type="button">
            Главная
          </button>
        </header>

        <main className="dashboardMain">
          <h1 className="dashboardTitle">Мои ссылки</h1>

          <section className="statsGrid">
            <article className="statCard">
              <span>Создано ссылок</span>
              <strong>{dashboardLinks.length}</strong>
            </article>

            <article className="statCard">
              <span>Всего переходов по ссылкам</span>
              <strong>{totalClicks}</strong>
            </article>

            <article className="statCard">
              <span>Максимальное число переходов по ссылке</span>
              <strong>{maxClicks}</strong>
            </article>
          </section>

          <section className="linksPanel">
            {dashboardLoading && (
              <div className="dashboardState">Загружаем ссылки...</div>
            )}

            {dashboardMessage && (
              <div className="dashboardMessage">{dashboardMessage}</div>
            )}

            {dashboardError && <div className="errorBox">{dashboardError}</div>}

            {!dashboardLoading && !dashboardError && dashboardLinks.length === 0 && (
              <div className="dashboardState">
                У вас пока нет сохранённых ссылок
              </div>
            )}

            {!dashboardLoading && !dashboardError && dashboardLinks.length > 0 && (
              <div className="linksTable">
                <div className="linksTableHead">
                  <span>Короткий URL</span>
                  <span>Оригинальный URL</span>
                  <span>Клики</span>
                  <span>Создано</span>
                  <span>Действия</span>
                </div>

                {dashboardLinks.map((link) => (
                  <div className="linkEntry" key={link.code}>
                    <div className="linksTableRow">
                      <a href={link.short_url} rel="noreferrer" target="_blank">
                        {link.short_url}
                      </a>
                      <span className="originalUrl">{link.original_url}</span>
                      <span>{link.clicks_count}</span>
                      <span>{formatDate(link.created_at)}</span>
                      <div className="rowActions">
                        <button
                          onClick={() =>
                            navigator.clipboard.writeText(link.short_url)
                          }
                          type="button"
                        >
                          Copy
                        </button>
                        <button onClick={() => toggleDashboardQr(link)} type="button">
                          QR
                        </button>
                        <button onClick={() => startEditing(link)} type="button">
                          Редактировать
                        </button>
                        <button
                          className="dangerButton"
                          disabled={deleteLoadingCode === link.code}
                          onClick={() => handleDelete(link)}
                          type="button"
                        >
                          {deleteLoadingCode === link.code ? "Удаляем..." : "Удалить"}
                        </button>
                      </div>
                    </div>

                    {editingCode === link.code && (
                      <div className="editRow">
                        <div className="editGrid">
                          <label className="field">
                            <span>Оригинальный URL</span>
                            <input
                              className="urlInput"
                              onChange={(e) => setEditOriginalUrl(e.target.value)}
                              value={editOriginalUrl}
                            />
                          </label>

                          <label className="field">
                            <span>Alias</span>
                            <input
                              className="urlInput"
                              onChange={(e) => setEditAlias(e.target.value)}
                              value={editAlias}
                            />
                          </label>
                        </div>

                        {editError && <div className="editError">{editError}</div>}

                        <div className="editActions">
                          <button
                            disabled={editLoading}
                            onClick={() => handleSaveEdit(link)}
                            type="button"
                          >
                            {editLoading ? "Сохраняем..." : "Сохранить"}
                          </button>
                          <button
                            className="mutedButton"
                            disabled={editLoading}
                            onClick={cancelEditing}
                            type="button"
                          >
                            Отмена
                          </button>
                        </div>
                      </div>
                    )}

                    {dashboardQrLoadingCode === link.code && (
                      <div className="qrStatus">Загружаем QR...</div>
                    )}

                    {dashboardQrError?.code === link.code && (
                      <div className="qrError">{dashboardQrError.message}</div>
                    )}

                    {dashboardQr?.code === link.code && (
                      <div className="dashboardQrBox">
                        <img
                          className="qrImage"
                          src={dashboardQr.dataUrl}
                          alt={`QR ${link.code}`}
                        />
                        <div className="qrText">
                          <p>{link.short_url}</p>
                          <a href={dashboardQr.dataUrl} download={`${link.code}.png`}>
                            <button type="button">Скачать QR</button>
                          </a>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </section>
        </main>
      </div>
    );
  }

  if (screen === "auth") {
    return (
      <div className="authPage">
        <div className="authShell">
          <Logo onClick={openHome} />

          <section className="authCard">
            <div className="authTabs">
              <button
                className={authMode === "register" ? "authTab active" : "authTab"}
                onClick={() => switchAuthMode("register")}
                disabled={authLoading}
                type="button"
              >
                Регистрация
              </button>

              <button
                className={authMode === "login" ? "authTab active" : "authTab"}
                onClick={() => switchAuthMode("login")}
                disabled={authLoading}
                type="button"
              >
                Вход
              </button>
            </div>

            {authError && <div className="authError">{authError}</div>}

            {authMode === "register" ? (
              <form className="authForm" onSubmit={handleAuthSubmit}>
                <label className="field">
                  <span>Имя пользователя</span>
                  <input
                    className="urlInput"
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="Панамкин"
                    value={username}
                  />
                </label>

                <label className="field">
                  <span>Почта</span>
                  <input
                    className="urlInput"
                    onChange={(e) => setAuthEmail(e.target.value)}
                    placeholder="you@example.com"
                    required
                    type="email"
                    value={authEmail}
                  />
                </label>

                <label className="field">
                  <span>Пароль</span>
                  <input
                    className="urlInput"
                    onChange={(e) => setAuthPassword(e.target.value)}
                    placeholder="Пароль"
                    required
                    type="password"
                    value={authPassword}
                  />
                </label>

                <button className="authSubmit" disabled={authLoading} type="submit">
                  {authLoading ? "Регистрируем..." : "Зарегистрироваться"}
                </button>
              </form>
            ) : (
              <form className="authForm" onSubmit={handleAuthSubmit}>
                <label className="field">
                  <span>Почта</span>
                  <input
                    className="urlInput"
                    onChange={(e) => setAuthEmail(e.target.value)}
                    placeholder="you@example.com"
                    required
                    type="email"
                    value={authEmail}
                  />
                </label>

                <label className="field">
                  <span>Пароль</span>
                  <input
                    className="urlInput"
                    onChange={(e) => setAuthPassword(e.target.value)}
                    placeholder="Пароль"
                    required
                    type="password"
                    value={authPassword}
                  />
                </label>

                <button
                  className="forgotLink"
                  onClick={() => alert("Восстановление пароля пока не реализовано")}
                  type="button"
                >
                  Забыли пароль?
                </button>

                <button className="authSubmit" disabled={authLoading} type="submit">
                  {authLoading ? "Входим..." : "Войти"}
                </button>
              </form>
            )}
          </section>
        </div>
      </div>
    );
  }

  return (
    <div className="page">
      <header className="header">
        <Logo onClick={openHome} />

        {user ? (
          <div className="userArea">
            <span className="userEmail">{user.email}</span>
            <button
              className="cabinetButton"
              onClick={openDashboard}
              type="button"
            >
              Кабинет
            </button>
            <button className="logoutButton" onClick={handleLogout} type="button">
              Выйти
            </button>
          </div>
        ) : (
          <button className="outlineButton" onClick={openLogin} type="button">
            Войти
          </button>
        )}
      </header>

      <main className="main">
        <section className="hero">
          <div className="heroCopy">
            <p className="eyebrow">Короткие ссылки и QR-коды</p>

            <h1>Сократите длинную ссылку за пару секунд</h1>

            <p className="description">
              Вставьте URL, задайте понятный alias при необходимости и получите
              короткую ссылку, которую удобно отправлять клиентам, коллегам и
              подписчикам.
            </p>
          </div>

          <section className="shortenerCard">
            <div className="formGrid">
              <label className="field">
                <span>Длинная ссылка</span>
                <input
                  className="urlInput"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/long/path"
                />
              </label>

              <label className="field">
                <span>Alias, необязательно</span>
                <input
                  className="urlInput"
                  value={customCode}
                  onChange={(e) => setCustomCode(e.target.value)}
                  placeholder="MyAlias123"
                />
              </label>
            </div>

            <button
              className="shortenButton"
              onClick={handleShorten}
              disabled={loading}
              type="button"
            >
              {loading ? "Сокращаем..." : "Сократить"}
            </button>

            {error && <div className="errorBox">{error}</div>}

            {shortUrl && (
              <div className="result">
                <label>Ваша короткая ссылка</label>

                <div className="resultRow">
                  <input value={shortUrl} readOnly />
                  <button
                    onClick={() => navigator.clipboard.writeText(shortUrl)}
                    type="button"
                  >
                    Копировать
                  </button>
                </div>

                <div className="divider" />

                {token && (
                  <p className="savedHint">Ссылка сохранена в личном кабинете</p>
                )}

                {qrLoading && <p className="qrStatus">Загружаем QR...</p>}
                {qrError && <p className="qrError">{qrError}</p>}

                {qrCode && (
                  <div className="qrRow">
                    <img src={qrCode} alt="QR Code" className="qrImage" />

                    <div className="qrText">
                      <p>Сканируй QR-код, чтобы открыть ссылку</p>

                      <a href={qrCode} download="qrcode.png">
                        <button type="button">Скачать QR</button>
                      </a>
                    </div>
                  </div>
                )}
              </div>
            )}
          </section>
        </section>
      </main>
    </div>
  );
}

export default App;
