import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Table } from 'antd';

interface Status {
  ip: string;
  ping_time: number;
  last_success_at: string;
}

const App: React.FC = () => {
  const [data, setData] = useState<Status[]>([]);

  useEffect(() => {
    fetchStatuses();
    const interval = setInterval(fetchStatuses, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchStatuses = async () => {
    try {
      const response = await axios.get<Status[]>('http://localhost:8080/statuses');
      setData(response.data);
    } catch (error) {
      console.error('Ошибка получения данных', error);
    }
  };

  const columns = [
    { title: 'IP адрес', dataIndex: 'ip', key: 'ip' },
    { title: 'Время пинга (мс)', dataIndex: 'ping_time', key: 'ping_time' },
    { title: 'Дата последней успешной попытки', dataIndex: 'last_success_at', key: 'last_success_at' }
  ];

  return (
    <div style={{ padding: 24 }}>
      <h1>Статус контейнеров</h1>
      <Table dataSource={data} columns={columns} rowKey="ip" />
    </div>
  );
};

export default App;
