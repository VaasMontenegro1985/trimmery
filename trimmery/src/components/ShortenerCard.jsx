import ResultBlock from "./ResultBlock";

function ShortenerCard({
  url,
  customCode,
  loading,
  error,
  shortUrl,
  qrCode,
  qrLoading,
  qrError,
  hasToken,
  onUrlChange,
  onCustomCodeChange,
  onShorten,
}) {
  return (
    <section className="shortenerCard">
      <div className="formGrid">
        <label className="field">
          <span>Длинная ссылка</span>
          <input
            className="urlInput"
            value={url}
            onChange={(e) => onUrlChange(e.target.value)}
            placeholder="https://example.com/long/path"
          />
        </label>

        <label className="field">
          <span>Алиас (необязательно)</span>
          <input
            className="urlInput"
            value={customCode}
            onChange={(e) => onCustomCodeChange(e.target.value)}
            placeholder="MyAlias123"
          />
        </label>
      </div>

      <button
        className="shortenButton"
        onClick={onShorten}
        disabled={loading}
        type="button"
      >
        {loading ? "Сокращаем..." : "Сократить"}
      </button>

      {error && <div className="errorBox">{error}</div>}

      <ResultBlock
        shortUrl={shortUrl}
        hasToken={hasToken}
        qrCode={qrCode}
        qrLoading={qrLoading}
        qrError={qrError}
      />
    </section>
  );
}

export default ShortenerCard;
