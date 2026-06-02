import { useEffect, useRef, useState } from "react";
import "./App.css";
import QRCode from "qrcode";
import {
  deleteLink,
  getLinkQrBlob,
  getLinkStats,
  getUserLinks,
  isValidToken,
  loginUser,
  registerUser,
  SESSION_EXPIRED_MESSAGE,
  shortenUrl,
  updateLink,
} from "./api/urlApi";
import AuthPage from "./components/AuthPage";
import DashboardPage from "./components/DashboardPage";
import HomePage from "./components/HomePage";

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
  const [dashboardStats, setDashboardStats] = useState(null);
  const [dashboardStatsLoadingCode, setDashboardStatsLoadingCode] = useState("");
  const [dashboardStatsError, setDashboardStatsError] = useState(null);
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
    setDashboardStats(null);
    setDashboardStatsLoadingCode("");
    setDashboardStatsError(null);
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
      clearSession(SESSION_EXPIRED_MESSAGE);
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
      const activeToken =
        isValidToken(token) && !isTokenExpired(expiresAt) ? token : null;

      if (isValidToken(token) && !activeToken) {
        clearSession(SESSION_EXPIRED_MESSAGE);
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
          if (err.message === SESSION_EXPIRED_MESSAGE) {
            clearSession(err.message);
          }

          setQrError("Не удалось загрузить QR-код");
        } finally {
          setQrLoading(false);
        }
      } else {
        const qr = await QRCode.toDataURL(data.short_url);
        replaceQrCode(qr);
      }
    } catch (err) {
      if (err.message === SESSION_EXPIRED_MESSAGE) {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

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
      if (err.message === SESSION_EXPIRED_MESSAGE) {
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
    setDashboardStats(null);
    setDashboardStatsError(null);
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
      if (err.message === SESSION_EXPIRED_MESSAGE) {
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

  const toggleDashboardStats = async (link) => {
    if (!hasActiveSession()) {
      setAuthMode("login");
      setScreen("auth");
      return;
    }

    if (dashboardStats?.code === link.code) {
      setDashboardStats(null);
      setDashboardStatsError(null);
      return;
    }

    setDashboardStats(null);
    setDashboardStatsError(null);
    setDashboardStatsLoadingCode(link.code);

    try {
      const stats = await getLinkStats(link.code, token, 10);
      setDashboardStats(stats);
    } catch (err) {
      if (err.message === SESSION_EXPIRED_MESSAGE) {
        clearSession(err.message);
        setAuthMode("login");
        setScreen("auth");
        return;
      }

      setDashboardStatsError({
        code: link.code,
        message: err.message,
      });
    } finally {
      setDashboardStatsLoadingCode("");
    }
  };

  const startEditing = (link) => {
    setEditingCode(link.code);
    setEditOriginalUrl(link.original_url);
    setEditAlias(link.code);
    setEditError("");
    setDashboardMessage("");
    setDashboardStats(null);
    setDashboardStatsError(null);
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
      if (err.message === SESSION_EXPIRED_MESSAGE) {
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

      if (dashboardStats?.code === link.code) {
        setDashboardStats(null);
        setDashboardStatsError(null);
      }

      if (dashboardStatsError?.code === link.code) {
        setDashboardStatsError(null);
      }

      if (dashboardStatsLoadingCode === link.code) {
        setDashboardStatsLoadingCode("");
      }

      if (editingCode === link.code) {
        cancelEditing();
      }

      await loadDashboardLinks(token);
      setDashboardMessage("Ссылка удалена");
    } catch (err) {
      if (err.message === SESSION_EXPIRED_MESSAGE) {
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
      <DashboardPage
        links={dashboardLinks}
        loading={dashboardLoading}
        error={dashboardError}
        dashboardMessage={dashboardMessage}
        totalClicks={totalClicks}
        maxClicks={maxClicks}
        dashboardQr={dashboardQr}
        dashboardQrLoadingCode={dashboardQrLoadingCode}
        dashboardQrError={dashboardQrError}
        dashboardStats={dashboardStats}
        dashboardStatsLoadingCode={dashboardStatsLoadingCode}
        dashboardStatsError={dashboardStatsError}
        editingCode={editingCode}
        editOriginalUrl={editOriginalUrl}
        editAlias={editAlias}
        editLoading={editLoading}
        editError={editError}
        deleteLoadingCode={deleteLoadingCode}
        onHome={openHome}
        onToggleQr={toggleDashboardQr}
        onToggleStats={toggleDashboardStats}
        onStartEditing={startEditing}
        onEditOriginalUrlChange={setEditOriginalUrl}
        onEditAliasChange={setEditAlias}
        onSaveEdit={handleSaveEdit}
        onCancelEditing={cancelEditing}
        onDelete={handleDelete}
      />
    );
  }

  if (screen === "auth") {
    return (
      <AuthPage
        authMode={authMode}
        authEmail={authEmail}
        authPassword={authPassword}
        username={username}
        authLoading={authLoading}
        authError={authError}
        onSwitchAuthMode={switchAuthMode}
        onUsernameChange={setUsername}
        onAuthEmailChange={setAuthEmail}
        onAuthPasswordChange={setAuthPassword}
        onSubmit={handleAuthSubmit}
        onHome={openHome}
      />
    );
  }

  return (
    <HomePage
      user={user}
      hasToken={Boolean(token)}
      url={url}
      customCode={customCode}
      loading={loading}
      error={error}
      shortUrl={shortUrl}
      qrCode={qrCode}
      qrLoading={qrLoading}
      qrError={qrError}
      onHome={openHome}
      onLogin={openLogin}
      onDashboard={openDashboard}
      onLogout={handleLogout}
      onUrlChange={setUrl}
      onCustomCodeChange={setCustomCode}
      onShorten={handleShorten}
    />
  );
}

export default App;
