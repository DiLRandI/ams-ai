import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import { PageHeader } from '../../components/PageHeader';
import { EmptyState } from '../../components/EmptyState';
import { Loading } from '../../components/Loading';
import { ReminderBadge } from '../../components/StateBadge';
import { dateOnly } from '../assets/format';

export function RemindersPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['reminders'], queryFn: api.reminders });

  if (isLoading) return <Loading />;
  if (error) return <div className="alert">{error instanceof Error ? error.message : 'Could not load reminders'}</div>;

  return (
    <>
      <PageHeader title="Reminders" eyebrow="30-day in-app reminder window" />
      <section className="panel">
        {!data?.length ? (
          <EmptyState title="No reminders" body="Expiry and service reminders appear here when dates enter the MVP reminder window." />
        ) : (
          <div className="tableWrap">
            <table>
              <thead>
                <tr>
                  <th>Due date</th>
                  <th>Reminder</th>
                  <th>Type</th>
                  <th>State</th>
                </tr>
              </thead>
              <tbody>
                {data.map((item) => (
                  <tr key={item.id}>
                    <td>{dateOnly(item.dueDate)}</td>
                    <td>
                      <Link to={`/assets/${item.assetId}`}>{item.title}</Link>
                      <span className="tableSub">{item.assetCode}</span>
                    </td>
                    <td>{item.sourceType}</td>
                    <td>
                      <ReminderBadge state={item.state} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  );
}
