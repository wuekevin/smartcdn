import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface User {
  id: number; username: string; email: string; is_active: boolean; created_at: string;
}

export default function UsersPage() {
  const qc = useQueryClient();
  const myId = Number(localStorage.getItem('user_id') ?? 0);
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => client.get('/users').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: { username: string; password: string; email?: string }) =>
      client.post('/users', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] });
      Message.success('用户已创建');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message;
      Message.error(msg || '创建失败');
    },
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
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>用户管理</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['users'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加用户</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加用户"
        visible={visible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical" autoComplete="off">
          <Form.Item label="用户名" field="username" rules={[{ required: true, minLength: 3, message: '用户名至少 3 个字符' }]}>
            <Input placeholder="用户名" autoComplete="off" />
          </Form.Item>
          <Form.Item label="密码" field="password" rules={[{ required: true, minLength: 6, message: '密码至少 6 个字符' }]}>
            <Input.Password placeholder="密码" autoComplete="new-password" />
          </Form.Item>
          <Form.Item label="邮箱（可选）" field="email">
            <Input placeholder="admin@example.com" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
