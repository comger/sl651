import React, { useState, useEffect } from 'react';
import { Layout, Table, Checkbox, Card, Typography, Space, Tag, theme, Divider, Empty, ConfigProvider } from 'antd';
import { DatabaseOutlined, HddOutlined, SyncOutlined } from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

// Configure axios base URL - Pointing to port 8085 as per config
const api = axios.create({
  baseURL: 'http://127.0.0.1:8085/api/v1',
});

const App = () => {
  const [devices, setDevices] = useState([]);
  const [selectedStationIDs, setSelectedStationIDs] = useState([]);
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [limit, setLimit] = useState(50);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Fetch devices on mount
  useEffect(() => {
    fetchDevices();
    const interval = setInterval(fetchDevices, 30000);
    return () => clearInterval(interval);
  }, []);

  // Poll data based on selection
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, refreshInterval * 1000);
    return () => clearInterval(interval);
  }, [selectedStationIDs, limit, refreshInterval]);

  const fetchDevices = async () => {
    try {
      const res = await api.get('/devices');
      if (res.data.code === 0) {
        const devList = res.data.data || [];
        setDevices(devList);
        // Auto-select all if selection is empty (first load)
        if (selectedStationIDs.length === 0 && devList.length > 0) {
          setSelectedStationIDs(devList.map(d => d.id));
        }
      }
    } catch (error) {
      console.error('Failed to fetch devices:', error);
    }
  };

  const fetchData = async () => {
    try {
      if (selectedStationIDs.length === 0) {
        // No need to fetch if empty, but we can set data to empty to be safe
        setData([]);
        return;
      }
      const params = { limit: limit };
      if (selectedStationIDs.length > 0) {
        params.station_ids = selectedStationIDs.join(',');
      }
      const res = await api.get('/data', { params });
      if (res.data.code === 0) {
        const rawData = res.data.data || [];
        // Ensure values is parsed if it's a string
        const processedData = rawData.map(item => {
          if (typeof item.values === 'string') {
            try {
              item.values = JSON.parse(item.values);
            } catch (e) {
              item.values = [];
            }
          }
          return item;
        });
        setData(processedData);
      }
    } catch (error) {
      console.error('Failed to fetch data:', error);
    }
  };

  const handleDeviceChange = (checkedValues) => {
    setSelectedStationIDs(checkedValues);
  };

  const columns = [
    {
      title: '采集时间',
      dataIndex: 'values',
      key: 'send_time',
      width: 180,
      render: (values) => {
        if (!Array.isArray(values)) return '-';
        const sendTime = values.find(v => v.tag === 'send_time');
        return sendTime && sendTime.value && sendTime.value.string ? sendTime.value.string : '-';
      },
    },
    {
      title: '入库时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (text) => moment(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '设备站码',
      dataIndex: 'device_id',
      key: 'device_id',
      width: 120,
      render: (text) => <Tag color="var(--ant-primary-color)">{text}</Tag>,
    },
    {
      title: '报文内容 & 解析结果',
      key: 'content',
      render: (_, record) => {
        // Prepare Raw Data
        const rawText = record.raw_data || '-';

        // Prepare Parsed Data
        const values = record.values;
        let parsedTags = null;
        if (Array.isArray(values)) {
          const keyMapping = {
            'water_level': '水位',
            'voltage': '电压',
            'cumulative_flow': '累计流量',
            'total_rainfall': '总雨量',
            'rainfall': '雨量',
            'instant_flow': '瞬时流量',
            'hourly_rainfall': '小时雨量',
            'day_rainfall': '日雨量',
            'period_rainfall': '时段雨量',
            'extra_rainfall': '额外雨量',
          };
          parsedTags = (
            <Space size={[4, 4]} wrap style={{ marginTop: 8 }}>
              {values.map((kp, index) => {
                const label = keyMapping[kp.tag] || kp.tag;
                if (kp.tag === 'station_id' || kp.tag === 'send_time') return null;
                if (!kp.value) return null;

                let val = kp.value.float ?? kp.value.int ?? kp.value.string ?? kp.value.bool;
                if (typeof val === 'number' && !Number.isInteger(val)) {
                  val = val.toFixed(3);
                }
                return (
                  <Tag key={index} color="blue" style={{ margin: 0, fontSize: '12px' }}>
                    {label}: {val}
                  </Tag>
                );
              })}
            </Space>
          );
        }

        return (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
            <Text code copyable style={{ maxWidth: '100%', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{rawText}</Text>
            {parsedTags}
          </div>
        );
      },
    },
  ];

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ display: 'flex', alignItems: 'center', background: '#001529', padding: '0 20px' }}>
          <DatabaseOutlined style={{ color: '#fff', fontSize: '24px', marginRight: '10px' }} />
          <Title level={3} style={{ color: '#fff', margin: 0 }}>SL651 监测与解析平台</Title>
        </Header>
        <Layout hasSider>
          <Sider width={280} style={{ background: colorBgContainer, padding: '20px', borderRight: '1px solid #f0f0f0', overflowY: 'auto' }}>
            <div style={{ marginBottom: '20px', display: 'flex', alignItems: 'center' }}>
              <HddOutlined style={{ marginRight: '8px', fontSize: '18px' }} />
              <Title level={4} style={{ margin: 0 }}>设备分组管家</Title>
            </div>
            <Divider style={{ margin: '10px 0' }} orientation="left">在线设备列表</Divider>
            {devices.length > 0 ? (
              <Checkbox.Group
                style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}
                value={selectedStationIDs}
                onChange={handleDeviceChange}
              >
                {devices.map(dev => (
                  <Card key={dev.id} size="small" hoverable style={{ borderLeft: '4px solid #1890ff' }} bodyStyle={{ padding: '8px' }}>
                    <Checkbox value={dev.id} style={{ width: '100%', marginLeft: 0 }}>
                      <Space direction="vertical" size={0} style={{ width: '100%' }}>
                        <Text strong>{dev.name || '未命名设备'}</Text>
                        <Text type="secondary" style={{ fontSize: '12px' }}>ID: {dev.id}</Text>
                      </Space>
                    </Checkbox>
                  </Card>
                ))}
              </Checkbox.Group>
            ) : (
              <Empty description="暂无在线设备" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Sider>
          <Content style={{ padding: '24px', overflow: 'hidden', background: '#f0f2f5', display: 'flex', flexDirection: 'column' }}>
            <Card
              bordered={false}
              style={{ borderRadius: borderRadiusLG, boxShadow: '0 1px 2px 0 rgba(0,0,0,0.03)', flex: 1, display: 'flex', flexDirection: 'column' }}
              bodyStyle={{ padding: '0', flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}
            >
              <div style={{ padding: '16px', borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Space>
                  <SyncOutlined spin={loading} />
                  <Title level={5} style={{ margin: 0 }}>实时报文</Title>
                  {selectedStationIDs.length > 0 ? <Tag color="processing">筛选: {selectedStationIDs.length} 个设备</Tag> : <Tag>显示全部</Tag>}
                </Space>
                <Space size="large">
                  <Space>
                    <Text>日志条数:</Text>
                    <input
                      type="range"
                      min="10"
                      max="100"
                      value={limit}
                      onChange={(e) => setLimit(parseInt(e.target.value))}
                      style={{ width: '100px' }}
                    />
                    <Tag>{limit}</Tag>
                  </Space>
                  <Space>
                    <Text>刷新频率:</Text>
                    <select
                      value={refreshInterval}
                      onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                      style={{ padding: '4px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
                    >
                      <option value={5}>5秒</option>
                      <option value={10}>10秒</option>
                      <option value={30}>30秒</option>
                      <option value={60}>60秒</option>
                    </select>
                  </Space>
                </Space>
              </div>
              <Table
                columns={columns}
                dataSource={data}
                rowKey="id"
                pagination={false}
                size="middle"
                style={{ flex: 1, overflow: 'hidden' }}
                scroll={{ y: 'calc(100vh - 200px)' }}
              />
            </Card>
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};

export default App;
