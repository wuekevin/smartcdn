import { Button, Card, Table, Tag, Typography } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

export default function CandidatesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['ip-candidates'],
    queryFn: () => client.get('/ip-candidates').then(r => r.data),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: 'IP', dataIndex: 'ip' },
    { title: 'ISP', dataIndex: 'isp' },
    { title: '域名 ID', dataIndex: 'domain_id', width: 90 },
    {
      title: '评分', dataIndex: 'score',
      render: (v: number) => v?.toFixed(2) ?? '-',
    },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={v === 'active' ? 'green' : v === 'failed' ? 'red' : 'gray'}>{v ?? '-'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>IP 候选</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['ip-candidates'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
