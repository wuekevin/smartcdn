import { Card, Table, Typography } from '@arco-design/web-react';
import { useQuery } from '@tanstack/react-query';
import client from '../../api/client';

export default function MetricsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['metrics'],
    queryFn: () => client.get('/metrics', { params: { limit: 50 } }).then(r => r.data),
    refetchInterval: 30_000,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '来源类型', dataIndex: 'source_type' },
    { title: '来源 ID', dataIndex: 'source_id', width: 90 },
    { title: '指标类型', dataIndex: 'metric_type' },
    {
      title: '值', dataIndex: 'value',
      render: (v: number) => v?.toFixed(4) ?? '-',
    },
    { title: '采集时间', dataIndex: 'collected_at', width: 180 },
  ];

  return (
    <Card title={<Typography.Title heading={6} style={{ margin: 0 }}>指标大盘</Typography.Title>}>
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
