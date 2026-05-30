import Logo from "./Logo";

function Header({ user, onHome, onLogin, onDashboard, onLogout }) {
  return (
    <header className="header">
      <Logo onClick={onHome} />

      {user ? (
        <div className="userArea">
          <span className="userEmail">{user.email}</span>
          <button className="cabinetButton" onClick={onDashboard} type="button">
            Кабинет
          </button>
          <button className="logoutButton" onClick={onLogout} type="button">
            Выйти
          </button>
        </div>
      ) : (
        <button className="outlineButton" onClick={onLogin} type="button">
          Войти
        </button>
      )}
    </header>
  );
}

export default Header;
