import {
  Button, Card, Table, Tag, Typography, Space, Tooltip, Message,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Certificate {
  id: number;
  domain_name: string;
  status: string;
  acme_provider: string;
  not_after: string | null;
  auto_renew: boolean;
}

function daysUntil(dateStr: string | null): number {
  if (!dateStr) return -1;
  return Math.floor((new Date(dateStr).getTime() - Date.now()) / 86_400_000);
}

function expiryColor(days: number): string {
  if (days < 0) return 'red';
  if (days < 7) return 'red';
  if (days < 30) return 'orange';
  return 'green';
}

export default function CertificatesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['certificates'],
    queryFn: () => client.get('/certificates').then((r) => r.data),
  });

  const retriggerMut = useMutation({
    mutationFn: (id: number) => client.post(`/certificates/${id}/retrigger`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['certificates'] }); Message.success('已触发重签'); },
    onError: () => Message.error('操作失败'),
  });

  const statusColor: Record<string, string> = {
    active: 'green', pending: 'orange', issuing: 'blue',
    failed: 'red', expired: 'red', revoked: 'gray',
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '域名', dataIndex: 'domain_name' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    { title: 'ACME 供应商', dataIndex: 'acme_provider' },
    {
      title: '到期', dataIndex: 'not_after',
      render: (v: string | null) => {
        const days = daysUntil(v);
        if (days < 0) return <Tag color="gray">未知</Tag>;
        return (
          <Tooltip content={v ?? ''}>
            <Tag color={expiryColor(days)}>{days} 天</Tag>
          </Tooltip>
        );
      },
    },
    {
      title: '自动续签', dataIndex: 'auto_renew',
      render: (v: boolean) => <Tag color={v ? 'arcoblue' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    {
      title: '操作', render: (_: unknown, row: Certificate) => (
        <Space>
          <Button size="mini" onClick={() => retriggerMut.mutate(row.id)}>重签</Button>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>SSL 证书</Typography.Title>}
      extra={
        <Space>
          <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['certificates'] })} />
          <Button type="primary" icon={<IconPlus />}>申请证书</Button>
        </Space>
      }
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
