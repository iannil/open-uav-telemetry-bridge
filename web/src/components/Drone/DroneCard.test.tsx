// Tests for DroneCard component
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DroneCard } from './DroneCard';
import { mockDroneState, mockDroneState2 } from '../../test/mocks';

describe('DroneCard', () => {
  it('should render drone device_id', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(mockDroneState.device_id)).toBeInTheDocument();
  });

  it('should render protocol source', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(mockDroneState.protocol_source)).toBeInTheDocument();
  });

  it('should render ARMED status when armed', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText('ARMED')).toBeInTheDocument();
  });

  it('should render DISARMED status when not armed', () => {
    render(<DroneCard drone={mockDroneState2} />);

    expect(screen.getByText('DISARMED')).toBeInTheDocument();
  });

  it('should render position coordinates', () => {
    render(<DroneCard drone={mockDroneState} />);

    // Check for latitude and longitude formatted values
    expect(screen.getByText(/39\.90420/)).toBeInTheDocument();
    expect(screen.getByText(/116\.40740/)).toBeInTheDocument();
  });

  it('should render altitude', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(`${mockDroneState.location.alt_gnss.toFixed(1)}m`)).toBeInTheDocument();
  });

  it('should render calculated speed', () => {
    render(<DroneCard drone={mockDroneState} />);

    const expectedSpeed = Math.sqrt(
      mockDroneState.velocity.vx ** 2 + mockDroneState.velocity.vy ** 2
    ).toFixed(1);
    expect(screen.getByText(`${expectedSpeed} m/s`)).toBeInTheDocument();
  });

  it('should render battery percentage', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(`${mockDroneState.status.battery_percent}%`)).toBeInTheDocument();
  });

  it('should render flight mode', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(mockDroneState.status.flight_mode)).toBeInTheDocument();
  });

  it('should render satellite count', () => {
    render(<DroneCard drone={mockDroneState} />);

    expect(screen.getByText(`${mockDroneState.status.satellites_visible} sats`)).toBeInTheDocument();
  });

  it('should call onClick when clicked', () => {
    const handleClick = vi.fn();
    render(<DroneCard drone={mockDroneState} onClick={handleClick} />);

    const card = screen.getByText(mockDroneState.device_id).closest('div[class*="cursor-pointer"]');
    fireEvent.click(card!);

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('should apply selected styles when isSelected is true', () => {
    const { container } = render(<DroneCard drone={mockDroneState} isSelected={true} />);

    const card = container.querySelector('.bg-blue-900\\/50');
    expect(card).toBeInTheDocument();
  });

  it('should apply default styles when isSelected is false', () => {
    const { container } = render(<DroneCard drone={mockDroneState} isSelected={false} />);

    const card = container.querySelector('.bg-gray-800');
    expect(card).toBeInTheDocument();
  });

  describe('battery color', () => {
    it('should show green for battery > 50%', () => {
      const drone = { ...mockDroneState, status: { ...mockDroneState.status, battery_percent: 75 } };
      const { container } = render(<DroneCard drone={drone} />);

      const batteryText = container.querySelector('.text-green-500');
      expect(batteryText).toBeInTheDocument();
      expect(batteryText?.textContent).toBe('75%');
    });

    it('should show yellow for battery between 20-50%', () => {
      const drone = { ...mockDroneState, status: { ...mockDroneState.status, battery_percent: 35 } };
      const { container } = render(<DroneCard drone={drone} />);

      const batteryText = container.querySelector('.text-yellow-500');
      expect(batteryText).toBeInTheDocument();
      expect(batteryText?.textContent).toBe('35%');
    });

    it('should show red for battery <= 20%', () => {
      const drone = { ...mockDroneState, status: { ...mockDroneState.status, battery_percent: 15 } };
      const { container } = render(<DroneCard drone={drone} />);

      const batteryText = container.querySelector('.text-red-500');
      expect(batteryText).toBeInTheDocument();
      expect(batteryText?.textContent).toBe('15%');
    });
  });
});
