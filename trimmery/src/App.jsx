import { useState } from "react";
import "./App.css";
import QRCode from "qrcode";
import { shortenUrl } from "./api/urlApi";

function App() {
  const [url, setUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [qrCode, setQrCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleShorten = async () => {
    const inputValue = url.trim();

    if (!inputValue || loading) return;

    setLoading(true);
    setError("");

    try {
      const data = await shortenUrl(inputValue);

      setShortUrl(data.short_url);

      // генерим QR в base64 PNG
      const qr = await QRCode.toDataURL(data.short_url);
      setQrCode(qr);
    } catch (err) {
      setError(err.message);
      setShortUrl("");
      setQrCode("");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page">
      <header className="header">
        <div className="logo">
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
        </div>

        <button className="outlineButton">Войти</button>
      </header>

      <main className="main">
        <p className="description">
          Забудьте о длинных и некрасивых адресах! С короткой ссылкой клиенты найдут
          вашу страницу в один миг — без лишнего текста и визуального шума.
        </p>

        <section className="shortenerCard">
          <input
            className="urlInput"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="Вставьте ссылку, которую нужно сократить"
          />

          <button
            className="shortenButton"
            onClick={handleShorten}
            disabled={loading}
          >
            {loading ? "Сокращаем..." : "Сократить"}
          </button>

          {error && <div className="result">{error}</div>}

          {shortUrl && (
            <div className="result">
              <label>Ваша короткая ссылка</label>

              <div className="resultRow">
                <input value={shortUrl} readOnly />
                <button onClick={() => navigator.clipboard.writeText(shortUrl)}>
                  Copy
                </button>
              </div>

              <div className="divider" />

              <div className="qrRow">
                {qrCode && (
                  <img src={qrCode} alt="QR Code" className="qrImage" />
                )}

                <div className="qrText">
                  <p>Сканируй QR-код, чтобы открыть ссылку</p>

                  <a href={qrCode} download="qrcode.png">
                    <button>Скачать QR</button>
                  </a>
                </div>
              </div>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;
