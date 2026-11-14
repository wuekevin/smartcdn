import React, { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space,
  Modal, Form, Input, Select, Tabs, Spin,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface ProviderField {
  key: string; label: string; type: string;
  required?: boolean; placeholder?: string; help_text?: string;
  options?: { label: string; value: string }[];
}

interface ProviderDef {
  type: string; label: string; category: string;
  description?: string; available?: boolean; fields: ProviderField[];
}

interface Credential {
  id: number; name: string; provider_type: string; category: string; created_at: string;
}

const CATEGORY_TABS = [
  { key: '', label: '全部' },
  { key: 'dns', label: 'DNS 验证' },
  { key: 'deploy', label: '部署目标' },
  { key: 'acme_ca', label: 'ACME CA' },
  { key: 'notify', label: '通知' },
];

const CATEGORY_COLORS: Record<string, string> = {
  dns: 'blue', deploy: 'green', acme_ca: 'purple', notify: 'orange',
};

const CATEGORY_LABELS: Record<string, string> = {
  dns: 'DNS 验证', deploy: '部署', acme_ca: 'ACME CA', notify: '通知',
};

export default function CredentialsPage() {
  const qc = useQueryClient();
  const [activeCategory, setActiveCategory] = useState('');
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<Credential | null>(null);
  const [selectedProviderType, setSelectedProviderType] = useState('');
  const [createCategory, setCreateCategory] = useState('dns');
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  // 获取 Provider 目录（schema 驱动）
  const { data: providersData, isLoading: providersLoading } = useQuery({
    queryKey: ['ssl-providers'],
    queryFn: () => client.get('/ssl/providers').then(r => r.data),
  });

  const providers: ProviderDef[] = providersData?.items ?? [];

  // 获取凭据列表
  const { data, isLoading } = useQuery({
    queryKey: ['ssl-credentials', activeCategory],
    queryFn: () => client.get('/ssl/credentials', { params: { category: activeCategory } }).then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: { name: string; provider_type: string; category: string; credentials: Record<string, string> }) =>
      client.post('/ssl/credentials', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['ssl-credentials'] });
      Message.success('凭据已创建');
      setCreateVisible(false);
      createForm.resetFields();
      setSelectedProviderType('');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/ssl/credentials/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['ssl-credentials'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) =>
      client.put(`/ssl/credentials/${id}`, { name }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['ssl-credentials'] });
      Message.success('已更新');
      setEditVisible(false);
      setEditingRow(null);
    },
    onError: () => Message.error('更新失败'),
  });

  const handleCreate = async () => {
    const vals = await createForm.validate();
    const { name, provider_type, ...rest } = vals;
    const selectedDef = providers.find(p => p.type === provider_type);
    createMut.mutate({
      name, provider_type,
      category: selectedDef?.category ?? createCategory,
      credentials: rest as Record<string, string>,
    });
  };

  const handleEdit = (row: Credential) => {
    setEditingRow(row);
    editForm.setFieldsValue({ name: row.name });
    setEditVisible(true);
  };

  // 当前类别下可选的 Provider
  const filteredProviders = providers.filter(p =>
    createCategory === '' || p.category === createCategory
  );

  // 选中的 Provider 字段定义
  const selectedProvider = providers.find(p => p.type === selectedProviderType);

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    {
      title: '类别', dataIndex: 'category',
      render: (v: string) => <Tag color={CATEGORY_COLORS[v] ?? 'gray'}>{CATEGORY_LABELS[v] ?? v}</Tag>,
    },
    {
      title: 'Provider', dataIndex: 'provider_type',
      render: (v: string) => {
        const def = providers.find(p => p.type === v);
        return def?.label ?? v;
      },
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: Credential) => (
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此凭据？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>凭据中心</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['ssl-credentials'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>添加凭据</Button>
          </Space>
        }
      >
        <Tabs activeTab={activeCategory} onChange={setActiveCategory} style={{ marginBottom: 16 }}>
          {CATEGORY_TABS.map(t => <Tabs.TabPane key={t.key} title={t.label} />)}
        </Tabs>
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* 新建凭据 */}
      <Modal
        title="添加凭据"
        visible={createVisible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); createForm.resetFields(); setSelectedProviderType(''); }}
        unmountOnExit
        style={{ width: 560 }}
      >
        {providersLoading ? <Spin /> : (
          <Form form={createForm} layout="vertical" autoComplete="off">
            <Form.Item label="凭据名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
              <Input placeholder="如：我的 Cloudflare 账号" />
            </Form.Item>

            <Form.Item label="类别" style={{ marginBottom: 8 }}>
              <Select value={createCategory} onChange={v => { setCreateCategory(v); setSelectedProviderType(''); createForm.resetFields(['provider_type']); }}>
                <Select.Option value="dns">DNS 验证</Select.Option>
                <Select.Option value="deploy">部署目标</Select.Option>
                <Select.Option value="acme_ca">ACME CA</Select.Option>
                <Select.Option value="notify">通知</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item label="Provider" field="provider_type" rules={[{ required: true, message: '请选择 Provider' }]}>
              <Select
                placeholder="选择 Provider"
                onChange={v => setSelectedProviderType(v as string)}
                showSearch
                filterOption={(inputValue, option) => {
                  const val = (option as React.ReactElement<{ value?: string }>).props?.value;
                  const prov = filteredProviders.find(p => p.type === val);
                  return (prov?.label ?? String(val ?? '')).toLowerCase().includes(inputValue.toLowerCase());
                }}
              >
                {filteredProviders.map(p => (
                  <Select.Option key={p.type} value={p.type}>
                    {p.label}
                    {!p.available && <span style={{ marginLeft: 6, fontSize: 11, color: '#86909c' }}>（即将支持）</span>}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>

            {selectedProvider?.description && (
              <div style={{ marginBottom: 12, color: '#86909C', fontSize: 12 }}>
                {selectedProvider.description}
              </div>
            )}

            {/* 动态渲染 Provider 字段 */}
            {selectedProvider?.fields.map(f => (
              <Form.Item
                key={f.key}
                label={f.label}
                field={f.key}
                rules={f.required ? [{ required: true, message: `请输入 ${f.label}` }] : []}
                extra={f.help_text ? <span style={{ fontSize: 11, color: '#86909C' }}>{f.help_text}</span> : undefined}
              >
                {f.type === 'password' ? (
                  <Input.Password placeholder={f.placeholder} />
                ) : f.type === 'textarea' ? (
                  <Input.TextArea placeholder={f.placeholder} rows={4} style={{ fontFamily: 'monospace', fontSize: 12 }} />
                ) : f.type === 'select' ? (
                  <Select>
                    {f.options?.map(o => <Select.Option key={o.value} value={o.value}>{o.label}</Select.Option>)}
                  </Select>
                ) : (
                  <Input placeholder={f.placeholder} />
                )}
              </Form.Item>
            ))}
          </Form>
        )}
      </Modal>

      {/* 编辑凭据 */}
      <Modal
        title="编辑凭据"
        visible={editVisible}
        onOk={() => editForm.validate().then(vals => editingRow && updateMut.mutate({ id: editingRow.id, name: vals.name }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="凭据名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
