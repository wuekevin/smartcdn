import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNAccount {
  id: number; name: string; provider_type: string; is_active: boolean; created_at: string;
}

const CDN_PROVIDERS = [
  { label: 'Cloudflare', value: 'cloudflare', fields: ['api_token'] },
  { label: '阿里云 CDN', value: 'aliyun_cdn', fields: ['access_key_id', 'access_key_secret'] },
];

export default function CDNProvidersPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();
  const [selectedProvider, setSelectedProvider] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['cdn-accounts'],
    queryFn: () => client.get('/cdn/accounts').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: { name: string; provider_type: string; credentials: Record<string, string> }) =>
      client.post('/cdn/accounts', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cdn-accounts'] });
      Message.success('CDN 账户已创建');
      setVisible(false);
      form.resetFields();
      setSelectedProvider('');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/cdn/accounts/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cdn-accounts'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleCreate = async () => {
    const vals = await form.validate();
    const { name, provider_type, ...rest } = vals;
    createMut.mutate({ name, provider_type, credentials: rest });
  };

  const credFields = CDN_PROVIDERS.find(p => p.value === selectedProvider)?.fields ?? [];

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
      title: '操作', render: (_: unknown, row: CDNAccount) => (
        <Popconfirm title="确认删除此 CDN 账户？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>CDN 服务商</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['cdn-accounts'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加账户</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加 CDN 账户"
        visible={visible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); setSelectedProvider(''); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical" autoComplete="off">
          <Form.Item label="账户名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：我的 Cloudflare 账户" />
          </Form.Item>
          <Form.Item label="CDN 服务商" field="provider_type" rules={[{ required: true, message: '请选择服务商' }]}>
            <Select placeholder="选择服务商" onChange={v => setSelectedProvider(v as string)}>
              {CDN_PROVIDERS.map(p => <Select.Option key={p.value} value={p.value}>{p.label}</Select.Option>)}
            </Select>
          </Form.Item>
          {credFields.map(f => (
            <Form.Item key={f} label={f} field={f} rules={[{ required: true, message: `请输入 ${f}` }]}>
              <Input.Password placeholder={f} />
            </Form.Item>
          ))}
        </Form>
      </Modal>
    </>
  );
}
