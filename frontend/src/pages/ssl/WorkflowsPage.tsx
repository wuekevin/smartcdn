import { useState } from 'react';
import {
  Button, Card, Table, Typography, Space, Popconfirm, Message, Modal, Form, Input,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit, IconPlayArrow } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface WorkflowTemplate {
  id: number; name: string; description: string; node_count: number; graph_json: string; created_at: string;
}

const DEFAULT_GRAPH = JSON.stringify({
  nodes: [
    { id: 'start', type: 'start', label: 'Start' },
    { id: 'end', type: 'end', label: 'End' },
  ],
  edges: [{ from: 'start', to: 'end' }],
}, null, 2);

export default function WorkflowsPage() {
  const qc = useQueryClient();
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<WorkflowTemplate | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => client.get('/workflow-templates').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/workflow-templates', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workflow-templates'] });
      Message.success('工作流已创建');
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
      client.put(`/workflow-templates/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workflow-templates'] });
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
    mutationFn: (id: number) => client.delete(`/workflow-templates/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['workflow-templates'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const triggerMut = useMutation({
    mutationFn: (id: number) => client.post(`/workflow-templates/${id}/trigger`, {
      trigger_type: 'manual',
    }),
    onSuccess: () => { Message.success('工作流已触发'); },
    onError: () => Message.error('触发失败'),
  });

  const handleEdit = (row: WorkflowTemplate) => {
    setEditingRow(row);
    editForm.setFieldsValue({ description: row.description, graph_json: row.graph_json });
    setEditVisible(true);
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description' },
    { title: '节点数', dataIndex: 'node_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: WorkflowTemplate) => (
        <Space>
          <Button size="mini" icon={<IconPlayArrow />} onClick={() => triggerMut.mutate(row.id)}>触发</Button>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此工作流？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const templateForm = (f: ReturnType<typeof Form.useForm>[0], isCreate = false) => (
    <Form form={f} layout="vertical">
      {isCreate && (
        <Form.Item label="工作流名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
          <Input placeholder="如：SSL 自动续签" />
        </Form.Item>
      )}
      <Form.Item label="描述" field="description">
        <Input.TextArea placeholder="可选描述" />
      </Form.Item>
      <Form.Item label="Graph JSON" field="graph_json" rules={[{ required: true, message: '请输入 Graph JSON' }]} initialValue={isCreate ? DEFAULT_GRAPH : undefined}>
        <Input.TextArea rows={8} placeholder="工作流 DAG 定义（JSON 格式）" style={{ fontFamily: 'monospace', fontSize: 12 }} />
      </Form.Item>
    </Form>
  );

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>工作流</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['workflow-templates'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>创建工作流</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="创建工作流"
        visible={createVisible}
        onOk={() => createForm.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); createForm.resetFields(); }}
        unmountOnExit
        style={{ width: 600 }}
      >
        {templateForm(createForm, true)}
      </Modal>

      <Modal
        title="编辑工作流"
        visible={editVisible}
        onOk={() => editForm.validate().then(vals => editingRow && updateMut.mutate({ id: editingRow.id, ...vals }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
        style={{ width: 600 }}
      >
        {templateForm(editForm, false)}
      </Modal>
    </>
  );
}
