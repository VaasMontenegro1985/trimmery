import Header from "./Header";
import ShortenerCard from "./ShortenerCard";

function HomePage({
  user,
  hasToken,
  url,
  customCode,
  loading,
  error,
  shortUrl,
  qrCode,
  qrLoading,
  qrError,
  onHome,
  onLogin,
  onDashboard,
  onLogout,
  onUrlChange,
  onCustomCodeChange,
  onShorten,
}) {
  return (
    <div className="page">
      <Header
        user={user}
        onHome={onHome}
        onLogin={onLogin}
        onDashboard={onDashboard}
        onLogout={onLogout}
      />

      <main className="main">
        <section className="hero">
          <div className="heroCopy">
            <p className="eyebrow">Короткие ссылки и QR-коды</p>

            <h1>Сократите длинную ссылку за пару секунд</h1>

            <p className="description">
              Вставьте URL, задайте понятный алиас при необходимости и получите
              короткую ссылку, которую удобно отправлять клиентам, коллегам и
              подписчикам.
            </p>
          </div>

          <ShortenerCard
            url={url}
            customCode={customCode}
            loading={loading}
            error={error}
            shortUrl={shortUrl}
            qrCode={qrCode}
            qrLoading={qrLoading}
            qrError={qrError}
            hasToken={hasToken}
            onUrlChange={onUrlChange}
            onCustomCodeChange={onCustomCodeChange}
            onShorten={onShorten}
          />
        </section>
      </main>
    </div>
  );
}

export default HomePage;
