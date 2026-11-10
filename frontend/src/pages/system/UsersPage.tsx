import { Button, Card, Table, Tag, Typography, Popconfirm, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface User {
  id: number; username: string; email: string; role: string; is_active: boolean; created_at: string;
}

export default function UsersPage() {
  const qc = useQueryClient();
  const myId = Number(localStorage.getItem('user_id') ?? 0);

  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => client.get('/users').then(r => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/users/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['users'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '用户名', dataIndex: 'username' },
    { title: '邮箱', dataIndex: 'email' },
    { title: '角色', dataIndex: 'role' },
    {
      title: '状态', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '启用' : '禁用'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: User) => (
        <Popconfirm title="确认删除此用户？" onOk={() => deleteMut.mutate(row.id)} disabled={row.id === myId}>
          <Button size="mini" type="text" status="danger" disabled={row.id === myId}>删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>用户管理</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['users'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
