import { Button, Card, Table, Typography, Popconfirm, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface NodePool {
  id: number; name: string; description: string; node_count: number; created_at: string;
}

export default function NodePoolsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['node-pools'],
    queryFn: () => client.get('/node-pools').then(r => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/node-pools/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['node-pools'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description' },
    { title: '节点数', dataIndex: 'node_count', width: 80 },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: NodePool) => (
        <Popconfirm title="确认删除此节点池？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>节点池</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['node-pools'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
