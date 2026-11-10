import {
  Button, Card, Table, Tag, Typography, Message,
} from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Alert {
  id: number;
  rule_name: string;
  metric_type: string;
  trigger_value: number;
  threshold: number;
  severity: string;
  status: string;
  fired_at: string;
  resolved_at: string | null;
  message: string;
}

export default function AlertsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['alerts'],
    queryFn: () => client.get('/alerts', { params: { limit: 50 } }).then((r) => r.data),
    refetchInterval: 30_000,
  });

  const resolveMut = useMutation({
    mutationFn: (id: number) => client.post(`/alerts/${id}/resolve`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['alerts'] }); Message.success('已手动解除'); },
    onError: () => Message.error('操作失败'),
  });

  const severityColor: Record<string, string> = { critical: 'red', warning: 'orange', info: 'blue' };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '规则', dataIndex: 'rule_name' },
    { title: '指标', dataIndex: 'metric_type' },
    {
      title: '触发值 / 阈值', render: (_: unknown, row: Alert) =>
        <span>{row.trigger_value.toFixed(2)} / {row.threshold.toFixed(2)}</span>,
    },
    {
      title: '严重度', dataIndex: 'severity',
      render: (v: string) => <Tag color={severityColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={v === 'firing' ? 'red' : 'green'}>{v === 'firing' ? '告警中' : '已恢复'}</Tag>,
    },
    { title: '触发时间', dataIndex: 'fired_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: Alert) =>
        row.status === 'firing'
          ? <Button size="mini" onClick={() => resolveMut.mutate(row.id)}>手动解除</Button>
          : null,
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>告警管理</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['alerts'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
