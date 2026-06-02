import { useEffect, useState } from "react";
import { copyToClipboard } from "../utils/copyToClipboard";

function ResultBlock({ shortUrl, hasToken, qrCode, qrLoading, qrError }) {
  const [copyStatus, setCopyStatus] = useState("idle");

  useEffect(() => {
    if (copyStatus === "idle") return undefined;

    const timeoutId = window.setTimeout(() => {
      setCopyStatus("idle");
    }, 1800);

    return () => window.clearTimeout(timeoutId);
  }, [copyStatus]);

  const handleCopy = async () => {
    const copied = await copyToClipboard(shortUrl);
    setCopyStatus(copied ? "success" : "fail");
  };

  if (!shortUrl) return null;

  return (
    <div className="result">
      <label>Ваша короткая ссылка</label>

      <div className="resultRow">
        <input value={shortUrl} readOnly />
        <button onClick={handleCopy} type="button">
          {copyStatus === "success"
            ? "Скопировано"
            : copyStatus === "fail"
              ? "Не скопировано"
              : "Копировать"}
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
          <img src={qrCode} alt="QR-код" className="qrImage" />

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
