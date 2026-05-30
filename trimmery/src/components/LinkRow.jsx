function formatDate(value) {
  if (!value) return "";

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(value));
}

function LinkRow({
  link,
  dashboardQr,
  dashboardQrLoadingCode,
  dashboardQrError,
  editingCode,
  editOriginalUrl,
  editAlias,
  editLoading,
  editError,
  deleteLoadingCode,
  onToggleQr,
  onStartEditing,
  onEditOriginalUrlChange,
  onEditAliasChange,
  onSaveEdit,
  onCancelEditing,
  onDelete,
}) {
  const isEditing = editingCode === link.code;

  return (
    <div className="linkEntry">
      <div className="linksTableRow">
        <a href={link.short_url} rel="noreferrer" target="_blank">
          {link.short_url}
        </a>
        <span className="originalUrl">{link.original_url}</span>
        <span>{link.clicks_count}</span>
        <span>{formatDate(link.created_at)}</span>
        <div className="rowActions">
          <button
            onClick={() => navigator.clipboard.writeText(link.short_url)}
            type="button"
          >
            Copy
          </button>
          <button onClick={() => onToggleQr(link)} type="button">
            QR
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
              <span>Alias</span>
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
