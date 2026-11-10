import { Button, Card, Table, Tag, Typography, Popconfirm, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface DNSAccount {
  id: number;
  name: string;
  provider_type: string;
  is_active: boolean;
  created_at: string;
}

export default function DNSAccountsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['dns-accounts'],
    queryFn: () => client.get('/dns-accounts').then(r => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/dns-accounts/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['dns-accounts'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '服务商', dataIndex: 'provider_type' },
    {
      title: '状态', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '启用' : '禁用'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: DNSAccount) => (
        <Popconfirm title="确认删除此 DNS 账户？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>DNS 账户</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['dns-accounts'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
