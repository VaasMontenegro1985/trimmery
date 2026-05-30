function DashboardStats({ linksCount, totalClicks, maxClicks }) {
  return (
    <section className="statsGrid">
      <article className="statCard">
        <span>Создано ссылок</span>
        <strong>{linksCount}</strong>
      </article>

      <article className="statCard">
        <span>Всего переходов по ссылкам</span>
        <strong>{totalClicks}</strong>
      </article>

      <article className="statCard">
        <span>Максимальное число переходов по ссылке</span>
        <strong>{maxClicks}</strong>
      </article>
    </section>
  );
}

export default DashboardStats;
