import { Button, Card, Table, Typography, Space, Message } from '@arco-design/web-react';
import { IconRefresh, IconPlayArrow } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface WorkflowTemplate {
  id: number; name: string; description: string; node_count: number; created_at: string;
}

export default function WorkflowsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => client.get('/workflow-templates').then(r => r.data),
  });

  const triggerMut = useMutation({
    mutationFn: (id: number) => client.post(`/workflow-templates/${id}/trigger`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['workflow-templates'] }); Message.success('已触发'); },
    onError: () => Message.error('触发失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description' },
    { title: '节点数', dataIndex: 'node_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: WorkflowTemplate) => (
        <Button size="mini" icon={<IconPlayArrow />} onClick={() => triggerMut.mutate(row.id)}>触发</Button>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>工作流</Typography.Title>}
      extra={
        <Space>
          <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['workflow-templates'] })} />
        </Space>
      }
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
