import { useTranslation } from 'react-i18next';

function Dashboard() {
  const { t } = useTranslation();

  return (
    <div>
      <h1>{t('dashboard.title')}</h1>
      <p>{t('dashboard.welcome')}</p>
      <button onClick={() => console.log(t('dashboard.submit'))}>
        {t('dashboard.click')}
      </button>
    </div>
  );
}
