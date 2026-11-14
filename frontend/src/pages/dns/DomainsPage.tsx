import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message, Modal, Form, Input, Switch, Select,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface Domain {
  id: number; name: string; status: string; icp_filed: boolean; dns_account_id: number; created_at: string;
}
interface DNSAccount { id: number; name: string; }

export default function DomainsPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['domains'],
    queryFn: () => client.get('/domains', { params: { limit: 20 } }).then(r => r.data),
  });

  const { data: accountsData } = useQuery({
    queryKey: ['dns-accounts'],
    queryFn: () => client.get('/dns-accounts').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/domains', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['domains'] });
      Message.success('域名已添加');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '添加失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/domains/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['domains'] }); Message.success('域名已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const accounts: DNSAccount[] = accountsData?.items ?? [];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '域名', dataIndex: 'name' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={v === 'active' ? 'green' : 'orange'}>{v}</Tag>,
    },
    {
      title: 'ICP 备案', dataIndex: 'icp_filed',
      render: (v: boolean) => <Tag color={v ? 'arcoblue' : 'gray'}>{v ? '已备案' : '未备案'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: Domain) => (
        <Space>
          <Popconfirm title="确认删除此域名？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>域名列表</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['domains'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加域名</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加域名"
        visible={visible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="域名" field="name" rules={[{ required: true, message: '请输入域名' }]}>
            <Input placeholder="example.com" />
          </Form.Item>
          <Form.Item label="DNS 账户" field="account_id" rules={[{ required: true, message: '请选择 DNS 账户' }]}>
            <Select placeholder="选择 DNS 账户">
              {accounts.map(a => <Select.Option key={a.id} value={Number(a.id)}>{a.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="ICP 备案" field="icp_filed" triggerPropName="checked" initialValue={false}>
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
