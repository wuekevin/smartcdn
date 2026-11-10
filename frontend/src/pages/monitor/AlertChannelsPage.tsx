import { Button, Card, Table, Tag, Typography, Popconfirm, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface AlertChannel {
  id: number; name: string; type: string; enabled: boolean; created_at: string;
}

export default function AlertChannelsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['alert-channels'],
    queryFn: () => client.get('/alert-channels').then(r => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/alert-channels/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['alert-channels'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

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
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>通知渠道</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['alert-channels'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
