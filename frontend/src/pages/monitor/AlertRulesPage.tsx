import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message, Modal, Form, Input, Select, InputNumber, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface AlertRule {
  id: number; name: string; description: string; source_type: string; metric_type: string;
  condition: string; threshold: number; severity: string; is_active: boolean; created_at: string;
}

const severityColor: Record<string, string> = {
  critical: 'red', warning: 'orange', info: 'blue',
};

const CONDITIONS = ['>', '>=', '<', '<=', '==', '!='];
const SEVERITIES = [
  { label: '严重 (critical)', value: 'critical' },
  { label: '警告 (warning)', value: 'warning' },
  { label: '信息 (info)', value: 'info' },
];

export default function AlertRulesPage() {
  const qc = useQueryClient();
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<AlertRule | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['alert-rules'],
    queryFn: () => client.get('/alert-rules').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/alert-rules', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alert-rules'] });
      Message.success('告警规则已创建');
      setCreateVisible(false);
      createForm.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, ...body }: { id: number } & Record<string, unknown>) =>
      client.put(`/alert-rules/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['alert-rules'] });
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
    mutationFn: (id: number) => client.delete(`/alert-rules/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['alert-rules'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleEdit = (row: AlertRule) => {
    setEditingRow(row);
    editForm.setFieldsValue({
      name: row.name,
      description: row.description,
      condition: row.condition,
      threshold: row.threshold,
      severity: row.severity,
      is_active: row.is_active,
    });
    setEditVisible(true);
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '规则名称', dataIndex: 'name' },
    { title: '指标类型', dataIndex: 'metric_type' },
    { title: '条件', dataIndex: 'condition', width: 60 },
    { title: '阈值', dataIndex: 'threshold', width: 80 },
    {
      title: '等级', dataIndex: 'severity',
      render: (v: string) => <Tag color={severityColor[v] ?? 'gray'}>{v}</Tag>,
    },
    {
      title: '启用', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '是' : '否'}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: AlertRule) => (
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此告警规则？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const ruleForm = (formInstance: ReturnType<typeof Form.useForm>[0]) => (
    <Form form={formInstance} layout="vertical">
      <Form.Item label="规则名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
        <Input placeholder="如：CPU 过高告警" />
      </Form.Item>
      <Form.Item label="描述" field="description">
        <Input.TextArea placeholder="可选描述" />
      </Form.Item>
      <Form.Item label="数据源类型" field="source_type" rules={[{ required: true, message: '请输入数据源类型' }]}>
        <Input placeholder="如：node, domain, probe" />
      </Form.Item>
      <Form.Item label="指标类型" field="metric_type" rules={[{ required: true, message: '请输入指标类型' }]}>
        <Input placeholder="如：cpu_percent, latency_ms" />
      </Form.Item>
      <Form.Item label="条件" field="condition" rules={[{ required: true, message: '请选择条件' }]}>
        <Select placeholder="选择比较条件">
          {CONDITIONS.map(c => <Select.Option key={c} value={c}>{c}</Select.Option>)}
        </Select>
      </Form.Item>
      <Form.Item label="阈值" field="threshold">
        <InputNumber />
      </Form.Item>
      <Form.Item label="告警等级" field="severity" initialValue="warning">
        <Select>
          {SEVERITIES.map(s => <Select.Option key={s.value} value={s.value}>{s.label}</Select.Option>)}
        </Select>
      </Form.Item>
      <Form.Item label="检测窗口（分钟）" field="window_minutes" initialValue={5}>
        <InputNumber min={1} />
      </Form.Item>
      <Form.Item label="持续时间（分钟）" field="for_minutes" initialValue={1}>
        <InputNumber min={0} />
      </Form.Item>
    </Form>
  );

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>告警规则</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['alert-rules'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>创建规则</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* 新建弹窗 */}
      <Modal
        title="创建告警规则"
        visible={createVisible}
        onOk={() => createForm.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); createForm.resetFields(); }}
        unmountOnExit
        style={{ width: 560 }}
      >
        {ruleForm(createForm)}
      </Modal>

      {/* 编辑弹窗 */}
      <Modal
        title="编辑告警规则"
        visible={editVisible}
        onOk={() => editForm.validate().then(vals => editingRow && updateMut.mutate({ id: editingRow.id, ...vals }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
        style={{ width: 560 }}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="规则名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="描述" field="description">
            <Input.TextArea />
          </Form.Item>
          <Form.Item label="条件" field="condition" rules={[{ required: true }]}>
            <Select>
              {CONDITIONS.map(c => <Select.Option key={c} value={c}>{c}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="阈值" field="threshold">
            <InputNumber />
          </Form.Item>
          <Form.Item label="告警等级" field="severity">
            <Select>
              {SEVERITIES.map(s => <Select.Option key={s.value} value={s.value}>{s.label}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item label="启用" field="is_active" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
