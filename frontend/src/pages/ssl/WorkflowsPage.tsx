import { useState } from 'react';
import {
  Button, Card, Table, Typography, Space, Popconfirm, Message,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit, IconPlayArrow } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';
import WorkflowEditor from '../../components/WorkflowEditor';

interface WorkflowTemplate {
  id: number; name: string; description: string; node_count: number; graph_json: string; created_at: string;
}

export default function WorkflowsPage() {
  const qc = useQueryClient();
  const [editorVisible, setEditorVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<WorkflowTemplate | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => client.get('/workflow-templates').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/workflow-templates', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workflow-templates'] });
      Message.success('工作流已创建');
      setEditorVisible(false);
      setEditingRow(null);
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
      setEditorVisible(false);
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
    mutationFn: (id: number) => client.post(`/workflow-templates/${id}/trigger`, { trigger_type: 'manual' }),
    onSuccess: () => Message.success('工作流已触发'),
    onError: () => Message.error('触发失败'),
  });

  const handleSave = (name: string, graphJSON: string) => {
    if (editingRow) {
      updateMut.mutate({ id: editingRow.id, name, graph_json: graphJSON });
    } else {
      createMut.mutate({ name, graph_json: graphJSON });
    }
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
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => {
            setEditingRow(row);
            setEditorVisible(true);
          }}>编辑</Button>
          <Popconfirm title="确认删除此工作流？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>工作流</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['workflow-templates'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => { setEditingRow(null); setEditorVisible(true); }}>
              创建工作流
            </Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <WorkflowEditor
        visible={editorVisible}
        initialGraphJSON={editingRow?.graph_json}
        initialName={editingRow?.name ?? ''}
        onSave={handleSave}
        onClose={() => { setEditorVisible(false); setEditingRow(null); }}
      />
    </>
  );
}
