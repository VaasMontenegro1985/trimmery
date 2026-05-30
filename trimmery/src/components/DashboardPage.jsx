import DashboardStats from "./DashboardStats";
import LinksTable from "./LinksTable";
import Logo from "./Logo";

function DashboardPage({
  links,
  loading,
  error,
  dashboardMessage,
  totalClicks,
  maxClicks,
  dashboardQr,
  dashboardQrLoadingCode,
  dashboardQrError,
  editingCode,
  editOriginalUrl,
  editAlias,
  editLoading,
  editError,
  deleteLoadingCode,
  onHome,
  onToggleQr,
  onStartEditing,
  onEditOriginalUrlChange,
  onEditAliasChange,
  onSaveEdit,
  onCancelEditing,
  onDelete,
}) {
  return (
    <div className="page dashboardPage">
      <header className="header dashboardHeader">
        <Logo onClick={onHome} />

        <button className="outlineButton" onClick={onHome} type="button">
          Главная
        </button>
      </header>

      <main className="dashboardMain">
        <h1 className="dashboardTitle">Мои ссылки</h1>

        <DashboardStats
          linksCount={links.length}
          totalClicks={totalClicks}
          maxClicks={maxClicks}
        />

        <LinksTable
          links={links}
          loading={loading}
          error={error}
          dashboardMessage={dashboardMessage}
          dashboardQr={dashboardQr}
          dashboardQrLoadingCode={dashboardQrLoadingCode}
          dashboardQrError={dashboardQrError}
          editingCode={editingCode}
          editOriginalUrl={editOriginalUrl}
          editAlias={editAlias}
          editLoading={editLoading}
          editError={editError}
          deleteLoadingCode={deleteLoadingCode}
          onToggleQr={onToggleQr}
          onStartEditing={onStartEditing}
          onEditOriginalUrlChange={onEditOriginalUrlChange}
          onEditAliasChange={onEditAliasChange}
          onSaveEdit={onSaveEdit}
          onCancelEditing={onCancelEditing}
          onDelete={onDelete}
        />
      </main>
    </div>
  );
}

export default DashboardPage;
