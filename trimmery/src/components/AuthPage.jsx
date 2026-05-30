import Logo from "./Logo";

function AuthPage({
  authMode,
  authEmail,
  authPassword,
  username,
  authLoading,
  authError,
  onSwitchAuthMode,
  onUsernameChange,
  onAuthEmailChange,
  onAuthPasswordChange,
  onSubmit,
  onHome,
}) {
  return (
    <div className="authPage">
      <div className="authShell">
        <Logo onClick={onHome} />

        <section className="authCard">
          <div className="authTabs">
            <button
              className={authMode === "register" ? "authTab active" : "authTab"}
              onClick={() => onSwitchAuthMode("register")}
              disabled={authLoading}
              type="button"
            >
              Регистрация
            </button>

            <button
              className={authMode === "login" ? "authTab active" : "authTab"}
              onClick={() => onSwitchAuthMode("login")}
              disabled={authLoading}
              type="button"
            >
              Вход
            </button>
          </div>

          {authError && <div className="authError">{authError}</div>}

          {authMode === "register" ? (
            <form className="authForm" onSubmit={onSubmit}>
              <label className="field">
                <span>Имя пользователя</span>
                <input
                  className="urlInput"
                  onChange={(e) => onUsernameChange(e.target.value)}
                  placeholder="Панамкин"
                  value={username}
                />
              </label>

              <label className="field">
                <span>Почта</span>
                <input
                  className="urlInput"
                  onChange={(e) => onAuthEmailChange(e.target.value)}
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
                  onChange={(e) => onAuthPasswordChange(e.target.value)}
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
            <form className="authForm" onSubmit={onSubmit}>
              <label className="field">
                <span>Почта</span>
                <input
                  className="urlInput"
                  onChange={(e) => onAuthEmailChange(e.target.value)}
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
                  onChange={(e) => onAuthPasswordChange(e.target.value)}
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

export default AuthPage;
