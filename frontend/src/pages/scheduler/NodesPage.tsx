import {
  Button, Card, Table, Tag, Typography, Space, Message,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNNode {
  id: number;
  name: string;
  ip: string;
  region: string;
  isp: string;
  status: string;
  https_latency_ms: number;
  https_status_code: number;
  last_seen_at: string | null;
}

export default function NodesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['cdn-nodes'],
    queryFn: () => client.get('/nodes').then((r) => r.data),
    refetchInterval: 30_000,
  });

  const statusColor: Record<string, string> = {
    active: 'green', inactive: 'gray', maintenance: 'orange', offline: 'red',
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '节点名称', dataIndex: 'name' },
    { title: 'IP', dataIndex: 'ip', width: 150 },
    { title: '地区', dataIndex: 'region' },
    { title: 'ISP', dataIndex: 'isp' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: 'HTTPS 延迟',
      render: (_: unknown, row: CDNNode) =>
        row.https_latency_ms > 0
          ? <span style={{ color: row.https_latency_ms > 300 ? 'orange' : 'green' }}>{row.https_latency_ms} ms</span>
          : <Tag color="gray">-</Tag>,
    },
    { title: '最后在线', dataIndex: 'last_seen_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: CDNNode) => (
        <Space>
          <Button size="mini" onClick={() =>
            client.post(`/nodes/${row.id}/https-check`).then(() => {
              qc.invalidateQueries({ queryKey: ['cdn-nodes'] });
              Message.success('HTTPS 检测已触发');
            })
          }>HTTPS 检测</Button>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>CDN 节点</Typography.Title>}
      extra={
        <Space>
          <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['cdn-nodes'] })} />
          <Button type="primary" icon={<IconPlus />}>添加节点</Button>
        </Space>
      }
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
