import LinkRow from "./LinkRow";

function LinksTable({
  links,
  loading,
  error,
  dashboardMessage,
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
  return (
    <section className="linksPanel">
      {loading && <div className="dashboardState">Загружаем ссылки...</div>}

      {dashboardMessage && (
        <div className="dashboardMessage">{dashboardMessage}</div>
      )}

      {error && <div className="errorBox">{error}</div>}

      {!loading && !error && links.length === 0 && (
        <div className="dashboardState">У вас пока нет сохранённых ссылок</div>
      )}

      {!loading && !error && links.length > 0 && (
        <div className="linksTable">
          <div className="linksTableHead">
            <span>Короткий URL</span>
            <span>Оригинальный URL</span>
            <span>Клики</span>
            <span>Создано</span>
            <span>Действия</span>
          </div>

          {links.map((link) => (
            <LinkRow
              key={link.code}
              link={link}
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
              onToggleQr={onToggleQr}
              onToggleStats={onToggleStats}
              onStartEditing={onStartEditing}
              onEditOriginalUrlChange={onEditOriginalUrlChange}
              onEditAliasChange={onEditAliasChange}
              onSaveEdit={onSaveEdit}
              onCancelEditing={onCancelEditing}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </section>
  );
}

export default LinksTable;
