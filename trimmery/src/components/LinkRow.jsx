import { useEffect, useState } from "react";
import { copyToClipboard } from "../utils/copyToClipboard";

function formatDate(value) {
  if (!value) return "";

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(value));
}

function formatDateTime(value) {
  if (!value) return "";

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function LinkRow({
  link,
  dashboardQr,
  dashboardQrLoadingCode,
  dashboardQrError,
  dashboardStats,
  dashboardStatsLoadingCode,
  dashboardStatsError,
  editingCode,
  editOriginalUrl,
  editAlias,
  editLoading,
  editError,
  deleteLoadingCode,
  onToggleQr,
  onToggleStats,
  onStartEditing,
  onEditOriginalUrlChange,
  onEditAliasChange,
  onSaveEdit,
  onCancelEditing,
  onDelete,
}) {
  const isEditing = editingCode === link.code;
  const isStatsOpen = dashboardStats?.code === link.code;
  const visits = dashboardStats?.visits || [];
  const isStatsLoading = dashboardStatsLoadingCode === link.code;
  const statsError =
    dashboardStatsError?.code === link.code ? dashboardStatsError.message : "";
  const [copyStatus, setCopyStatus] = useState("idle");

  useEffect(() => {
    if (copyStatus === "idle") return undefined;

    const timeoutId = window.setTimeout(() => {
      setCopyStatus("idle");
    }, 1800);

    return () => window.clearTimeout(timeoutId);
  }, [copyStatus]);

  const handleCopy = async () => {
    const copied = await copyToClipboard(link.short_url);
    setCopyStatus(copied ? "success" : "fail");
  };

  return (
    <div className="linkEntry">
      <div className="linksTableRow">
        <a
          className="shortUrl"
          href={link.short_url}
          rel="noreferrer"
          target="_blank"
        >
          {link.short_url}
        </a>
        <span className="originalUrl">{link.original_url}</span>
        <span>{link.clicks_count}</span>
        <span>{formatDate(link.created_at)}</span>
        <div className="rowActions">
          <button onClick={handleCopy} type="button">
            {copyStatus === "success"
              ? "Скопировано"
              : copyStatus === "fail"
                ? "Не скопировано"
                : "Копировать"}
          </button>
          <button onClick={() => onToggleQr(link)} type="button">
            QR
          </button>
          <button onClick={() => onToggleStats(link)} type="button">
            Статистика
          </button>
          <button onClick={() => onStartEditing(link)} type="button">
            Редактировать
          </button>
          <button
            className="dangerButton"
            disabled={deleteLoadingCode === link.code}
            onClick={() => onDelete(link)}
            type="button"
          >
            {deleteLoadingCode === link.code ? "Удаляем..." : "Удалить"}
          </button>
        </div>
      </div>

      {isStatsLoading && (
        <div className="linkStatsBox">
          <p className="linkStatsStatus">Загружаем статистику...</p>
        </div>
      )}

      {statsError && (
        <div className="linkStatsBox">
          <p className="linkStatsError">{statsError}</p>
        </div>
      )}

      {isStatsOpen && (
        <div className="linkStatsBox">
          <div className="linkStatsSummary">
            <span>Всего переходов</span>
            <strong>{dashboardStats.clicks_count}</strong>
          </div>

          {visits.length === 0 ? (
            <p className="linkStatsEmpty">Переходов пока нет</p>
          ) : (
            <div className="visitsList">
              {visits.map((visit, index) => (
                <div className="visitItem" key={`${visit.visited_at}-${index}`}>
                  <div className="visitMeta">
                    <strong>{formatDateTime(visit.visited_at)}</strong>
                    <span>{visit.ip || "IP не определен"}</span>
                  </div>
                  <p>{visit.user_agent || "Браузер не определен"}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {isEditing && (
        <div className="editRow">
          <div className="editGrid">
            <label className="field">
              <span>Оригинальный URL</span>
              <input
                className="urlInput"
                onChange={(e) => onEditOriginalUrlChange(e.target.value)}
                value={editOriginalUrl}
              />
            </label>

            <label className="field">
              <span>Алиас</span>
              <input
                className="urlInput"
                onChange={(e) => onEditAliasChange(e.target.value)}
                value={editAlias}
              />
            </label>
          </div>

          {editError && <div className="editError">{editError}</div>}

          <div className="editActions">
            <button
              disabled={editLoading}
              onClick={() => onSaveEdit(link)}
              type="button"
            >
              {editLoading ? "Сохраняем..." : "Сохранить"}
            </button>
            <button
              className="mutedButton"
              disabled={editLoading}
              onClick={onCancelEditing}
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
  );
}

export default LinkRow;
