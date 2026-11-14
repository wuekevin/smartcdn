import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface AlertChannel {
  id: number; name: string; type: string; enabled: boolean; created_at: string;
}

const CHANNEL_TYPES = [
  { label: 'Webhook', value: 'webhook', fields: [{ key: 'url', label: 'Webhook URL', required: true }] },
  { label: '邮件', value: 'email', fields: [{ key: 'to', label: '收件邮箱', required: true }] },
  { label: '钉钉', value: 'dingtalk', fields: [{ key: 'webhook_url', label: '钉钉 Webhook URL', required: true }, { key: 'secret', label: '加签密钥', required: false }] },
  { label: 'Slack', value: 'slack', fields: [{ key: 'webhook_url', label: 'Slack Webhook URL', required: true }] },
  { label: '飞书', value: 'feishu', fields: [{ key: 'webhook_url', label: '飞书 Webhook URL', required: true }] },
];

export default function AlertChannelsPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [selectedType, setSelectedType] = useState('');
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => client.get('/alert-channels').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: { name: string; type: string; config: Record<string, string> }) =>
      client.post('/alert-channels', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alert-channels'] });
      Message.success('通知渠道已创建');
      setVisible(false);
      form.resetFields();
      setSelectedType('');
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/alert-channels/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['alert-channels'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleCreate = async () => {
    const vals = await form.validate();
    const { name, type, ...rest } = vals;
    createMut.mutate({ name, type, config: rest });
  };

  const configFields = CHANNEL_TYPES.find(t => t.value === selectedType)?.fields ?? [];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type' },
    {
      title: '状态', dataIndex: 'enabled',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '启用' : '禁用'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: AlertChannel) => (
        <Popconfirm title="确认删除此通知渠道？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>通知渠道</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['alert-channels'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加渠道</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加通知渠道"
        visible={visible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); setSelectedType(''); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="渠道名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：运维告警群" />
          </Form.Item>
          <Form.Item label="类型" field="type" rules={[{ required: true, message: '请选择类型' }]}>
            <Select placeholder="选择通知类型" onChange={v => { setSelectedType(v as string); }}>
              {CHANNEL_TYPES.map(t => <Select.Option key={t.value} value={t.value}>{t.label}</Select.Option>)}
            </Select>
          </Form.Item>
          {configFields.map(f => (
            <Form.Item
              key={f.key}
              label={f.label}
              field={f.key}
              rules={f.required ? [{ required: true, message: `请输入 ${f.label}` }] : []}
            >
              <Input placeholder={f.label} />
            </Form.Item>
          ))}
        </Form>
      </Modal>
    </>
  );
}
