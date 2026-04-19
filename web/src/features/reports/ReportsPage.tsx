import { FileDown } from 'lucide-react';
import { getToken } from '../../api/client';
import { PageHeader } from '../../components/PageHeader';

const reports = [
  ['Asset list', '/api/reports/assets.csv'],
  ['Warranty expiry report', '/api/reports/warranties.csv'],
  ['Vehicle renewal report', '/api/reports/vehicle-renewals.csv'],
  ['Service history report', '/api/reports/service-history.csv'],
  ['Fuel log export', '/api/reports/fuel-logs.csv']
] as const;

export function ReportsPage() {
  const token = getToken();
  return (
    <>
      <PageHeader title="Reports" eyebrow="CSV exports" />
      <section className="panel reportGrid">
        {reports.map(([label, path]) => (
          <a key={path} className="reportLink" href={`${path}?token=${token ?? ''}`} onClick={(event) => event.preventDefault()}>
            <FileDown size={20} />
            <span>{label}</span>
            <button
              className="secondaryButton"
              type="button"
              onClick={async () => {
                const response = await fetch(path, {
                  headers: token ? { Authorization: `Bearer ${token}` } : {}
                });
                const blob = await response.blob();
                const url = URL.createObjectURL(blob);
                const anchor = document.createElement('a');
                anchor.href = url;
                anchor.download = path.split('/').pop() ?? 'report.csv';
                anchor.click();
                URL.revokeObjectURL(url);
              }}
            >
              Download
            </button>
          </a>
        ))}
      </section>
    </>
  );
}
