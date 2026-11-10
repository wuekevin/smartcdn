import { Card, Table, Tag, Typography } from '@arco-design/web-react';
import { useQuery } from '@tanstack/react-query';
import client from '../../api/client';

const statusColor: Record<string, string> = {
  pending: 'orange', running: 'blue', success: 'green', failed: 'red',
};

export default function PublishJobsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['publish-jobs'],
    queryFn: () => client.get('/publish-jobs').then(r => r.data),
    refetchInterval: 30_000,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '域名 ID', dataIndex: 'domain_id', width: 100 },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: '试运行', dataIndex: 'dry_run',
      render: (v: boolean) => <Tag color={v ? 'orange' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    { title: '记录数', dataIndex: 'records_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ];

  return (
    <Card title={<Typography.Title heading={6} style={{ margin: 0 }}>发布任务</Typography.Title>}>
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
