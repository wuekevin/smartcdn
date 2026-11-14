import { Button, Card, Table, Tag, Typography } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

export default function ProbesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['probe-agents'],
    queryFn: () => client.get('/probe-agents').then(r => r.data),
    refetchInterval: 30_000,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: 'ISP', dataIndex: 'isp' },
    { title: '地区', dataIndex: 'region' },
    {
      title: '状态', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? 'online' : 'offline'}</Tag>,
    },
    { title: '最后在线', dataIndex: 'last_seen_at', width: 180 },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>探针管理</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['probe-agents'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
