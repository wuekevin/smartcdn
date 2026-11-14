import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNAccount {
  id: number; name: string; provider_type: string; is_active: boolean; created_at: string;
}

const CDN_PROVIDERS = [
  { label: 'Cloudflare', value: 'cloudflare', fields: ['api_token'] },
  { label: '阿里云 CDN', value: 'aliyun_cdn', fields: ['access_key_id', 'access_key_secret'] },
  { label: '腾讯云 EdgeOne', value: 'tencent_eo', fields: ['secret_id', 'secret_key', 'zone_id'] },
];

export default function CDNProvidersPage() {
  const qc = useQueryClient();
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<CDNAccount | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
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
      setCreateVisible(false);
      createForm.resetFields();
      setSelectedProvider('');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, name, is_active }: { id: number; name: string; is_active: boolean }) =>
      client.put(`/cdn/accounts/${id}`, { name, is_active }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cdn-accounts'] });
      Message.success('已更新');
      setEditVisible(false);
      editForm.resetFields();
      setEditingRow(null);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '更新失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/cdn/accounts/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cdn-accounts'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleCreate = async () => {
    const vals = await createForm.validate();
    const { name, provider_type, ...rest } = vals;
    createMut.mutate({ name, provider_type, credentials: rest });
  };

  const handleEdit = (row: CDNAccount) => {
    setEditingRow(row);
    editForm.setFieldsValue({ name: row.name, is_active: row.is_active });
    setEditVisible(true);
  };

  const handleUpdate = async () => {
    if (!editingRow) return;
    const vals = await editForm.validate();
    updateMut.mutate({ id: editingRow.id, name: vals.name, is_active: vals.is_active ?? true });
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
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此 CDN 账户？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
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
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>添加账户</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* 新建弹窗 */}
      <Modal
        title="添加 CDN 账户"
        visible={createVisible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); createForm.resetFields(); setSelectedProvider(''); }}
        unmountOnExit
      >
        <Form form={createForm} layout="vertical" autoComplete="off">
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

      {/* 编辑弹窗 */}
      <Modal
        title="编辑 CDN 账户"
        visible={editVisible}
        onOk={handleUpdate}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical" autoComplete="off">
          <Form.Item label="账户名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="账户名称" />
          </Form.Item>
          <Form.Item label="启用状态" field="is_active" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
