function ResultBlock({ shortUrl, hasToken, qrCode, qrLoading, qrError }) {
  if (!shortUrl) return null;

  return (
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

      {hasToken && (
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
  );
}

export default ResultBlock;
