import { useState } from 'react';
import {
  Button, Card, Table, Typography, Popconfirm, Message, Space, Modal, Form, Input, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface NodePool {
  id: number; name: string; description: string; node_count: number; is_active: boolean; created_at: string;
}

export default function NodePoolsPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<NodePool | null>(null);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['node-pools'],
    queryFn: () => client.get('/node-pools').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/node-pools', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['node-pools'] });
      Message.success('节点池已创建');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '创建失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, ...body }: { id: number } & Record<string, unknown>) =>
      client.put(`/node-pools/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['node-pools'] });
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
    mutationFn: (id: number) => client.delete(`/node-pools/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['node-pools'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const handleEdit = (row: NodePool) => {
    setEditingRow(row);
    editForm.setFieldsValue({ name: row.name, description: row.description, is_active: row.is_active });
    setEditVisible(true);
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description' },
    { title: '节点数', dataIndex: 'node_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: NodePool) => (
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此节点池？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>节点池</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['node-pools'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>创建节点池</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="创建节点池"
        visible={visible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="如：华南节点池" />
          </Form.Item>
          <Form.Item label="描述" field="description">
            <Input.TextArea placeholder="可选描述" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑节点池"
        visible={editVisible}
        onOk={() => editForm.validate().then(vals => editingRow && updateMut.mutate({ id: editingRow.id, ...vals }))}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="描述" field="description">
            <Input.TextArea />
          </Form.Item>
          <Form.Item label="启用" field="is_active" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
