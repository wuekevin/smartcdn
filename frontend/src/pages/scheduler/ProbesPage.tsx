import {
  Button, Card, Table, Tag, Typography, Space, Popconfirm, Message,
} from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

export default function ProbesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['probe-agents'],
    queryFn: () => client.get('/probe-agents').then(r => r.data),
    refetchInterval: 30_000,
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/probe-agents/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['probe-agents'] }); Message.success('探针已停用'); },
    onError: () => Message.error('操作失败'),
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '名称', dataIndex: 'name' },
    { title: 'ISP', dataIndex: 'isp' },
    { title: '地区', dataIndex: 'region' },
    {
      title: '状态', dataIndex: 'is_active',
      render: (v: boolean) => <Tag color={v ? 'green' : 'gray'}>{v ? '在线' : '离线'}</Tag>,
    },
    { title: '最后在线', dataIndex: 'last_seen_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: { id: number; is_active: boolean }) => (
        <Space>
          {row.is_active && (
            <Popconfirm title="确认停用此探针？（探针可重新注册激活）" onOk={() => deleteMut.mutate(row.id)}>
              <Button size="mini" type="text" status="danger">停用</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>探针管理</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['probe-agents'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
