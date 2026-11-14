import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface AlertChannel {
  id: number; name: string; type: string; enabled: boolean; created_at: string;
}

const TEMPLATE_PLACEHOLDER = `[{{.Severity}}] {{.RuleName}}
{{.Message}}
状态: {{.Status}} | 时间: {{.FiredAt}}
触发值: {{.TriggerValue}} | 阈值: {{.Threshold}}`;

const CHANNEL_TYPES = [
  {
    label: 'Telegram',
    value: 'telegram',
    fields: [
      { key: 'bot_token', label: 'Bot Token', required: true, placeholder: '从 @BotFather 获取' },
      { key: 'chat_id', label: 'Chat ID', required: true, placeholder: '群组或频道的 Chat ID' },
    ],
  },
  {
    label: '钉钉',
    value: 'dingtalk',
    fields: [
      { key: 'webhook_url', label: '钉钉 Webhook URL', required: true, placeholder: 'https://oapi.dingtalk.com/robot/send?access_token=...' },
      { key: 'secret', label: '加签密钥（可选）', required: false, placeholder: 'SEC...' },
    ],
  },
  {
    label: '飞书',
    value: 'feishu',
    fields: [
      { key: 'webhook_url', label: '飞书 Webhook URL', required: true, placeholder: 'https://open.feishu.cn/open-apis/bot/v2/hook/...' },
    ],
  },
  {
    label: 'Slack',
    value: 'slack',
    fields: [
      { key: 'webhook_url', label: 'Slack Webhook URL', required: true, placeholder: 'https://hooks.slack.com/services/...' },
    ],
  },
  {
    label: 'Webhook',
    value: 'webhook',
    fields: [
      { key: 'url', label: 'Webhook URL', required: true, placeholder: 'https://example.com/notify' },
    ],
  },
  {
    label: '邮件',
    value: 'email',
    fields: [
      { key: 'to', label: '收件邮箱', required: true, placeholder: 'ops@example.com' },
    ],
  },
];

const TYPE_LABELS: Record<string, string> = {
  telegram: 'Telegram', dingtalk: '钉钉', feishu: '飞书',
  slack: 'Slack', webhook: 'Webhook', email: '邮件',
};

const TYPE_COLORS: Record<string, string> = {
  telegram: 'blue', dingtalk: 'orange', feishu: 'cyan',
  slack: 'purple', webhook: 'gray', email: 'green',
};

export default function AlertChannelsPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<AlertChannel | null>(null);
  const [selectedType, setSelectedType] = useState('');
  const [showTemplate, setShowTemplate] = useState(false);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

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
      setShowTemplate(false);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, name, enabled }: { id: number; name: string; enabled: boolean }) =>
      client.put(`/alert-channels/${id}`, { name, enabled }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alert-channels'] });
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
    mutationFn: (id: number) => client.delete(`/alert-channels/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['alert-channels'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleEdit = (row: AlertChannel) => {
    setEditingRow(row);
    editForm.setFieldsValue({ name: row.name, enabled: row.enabled });
    setEditVisible(true);
  };

  const handleCreate = async () => {
    const vals = await form.validate();
    const { name, type, message_template, ...rest } = vals;
    const config: Record<string, string> = { ...rest };
    if (message_template) config.message_template = message_template;
    createMut.mutate({ name, type, config });
  };

  const configFields = CHANNEL_TYPES.find(t => t.value === selectedType)?.fields ?? [];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    {
      title: '类型', dataIndex: 'type',
      render: (v: string) => (
        <Tag color={TYPE_COLORS[v] ?? 'gray'}>
          {TYPE_LABELS[v] ?? v}
        </Tag>
      ),
    },
    {
      title: '状态', dataIndex: 'enabled',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '启用' : '禁用'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: AlertChannel) => (
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此通知渠道？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
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

      {/* Edit modal — only name + enabled (config is immutable after creation) */}
      <Modal
        title="编辑通知渠道"
        visible={editVisible}
        onOk={() => editForm.validate().then(vals => editingRow && updateMut.mutate({ id: editingRow.id, name: vals.name, enabled: vals.enabled ?? true }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="渠道名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="启用状态" field="enabled" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Create modal */}
      <Modal
        title="添加通知渠道"
        visible={visible}
        onOk={handleCreate}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); setSelectedType(''); setShowTemplate(false); }}
        unmountOnExit
        style={{ width: 520 }}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="渠道名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：运维告警群" />
          </Form.Item>
          <Form.Item label="类型" field="type" rules={[{ required: true, message: '请选择类型' }]}>
            <Select
              placeholder="选择通知类型"
              onChange={v => { setSelectedType(v as string); setShowTemplate(false); }}
            >
              {CHANNEL_TYPES.map(t => (
                <Select.Option key={t.value} value={t.value}>{t.label}</Select.Option>
              ))}
            </Select>
          </Form.Item>

          {/* Channel-specific fields */}
          {configFields.map(f => (
            <Form.Item
              key={f.key}
              label={f.label}
              field={f.key}
              rules={f.required ? [{ required: true, message: `请输入 ${f.label}` }] : []}
            >
              <Input placeholder={f.placeholder} />
            </Form.Item>
          ))}

          {/* Optional message template */}
          {selectedType && (
            <>
              <div style={{ marginBottom: 8 }}>
                <Button
                  type="text"
                  size="small"
                  onClick={() => setShowTemplate(v => !v)}
                  style={{ padding: 0, color: '#165DFF' }}
                >
                  {showTemplate ? '▼' : '▶'} 自定义消息模板（可选）
                </Button>
              </div>
              {showTemplate && (
                <Form.Item
                  field="message_template"
                  extra={
                    <span style={{ fontSize: 11, color: '#86909C', lineHeight: 1.6 }}>
                      支持变量：{'{{.RuleName}}'} {'{{.Severity}}'} {'{{.Status}}'} {'{{.Message}}'} {'{{.FiredAt}}'} {'{{.TriggerValue}}'} {'{{.Threshold}}'}
                    </span>
                  }
                >
                  <Input.TextArea
                    rows={4}
                    placeholder={TEMPLATE_PLACEHOLDER}
                    style={{ fontFamily: 'monospace', fontSize: 12 }}
                  />
                </Form.Item>
              )}
            </>
          )}
        </Form>
      </Modal>
    </>
  );
}
