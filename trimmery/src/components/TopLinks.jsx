function TopLinks({ links }) {
  const topLinks = [...links]
    .filter((link) => (link.clicks_count || 0) > 0)
    .sort((a, b) => (b.clicks_count || 0) - (a.clicks_count || 0))
    .slice(0, 3);

  return (
    <section className="topLinks">
      <div className="topLinksHeader">
        <h2>Топ ссылок</h2>
      </div>

      {topLinks.length === 0 ? (
        <p className="topLinksEmpty">Переходов пока нет</p>
      ) : (
        <div className="topLinksList">
          {topLinks.map((link, index) => (
            <a
              className="topLinkItem"
              href={link.short_url}
              key={link.code}
              rel="noreferrer"
              target="_blank"
            >
              <span className="topLinkRank">{index + 1}</span>
              <span className="topLinkUrl">{link.short_url}</span>
              <strong>{link.clicks_count}</strong>
            </a>
          ))}
        </div>
      )}
    </section>
  );
}

export default TopLinks;
