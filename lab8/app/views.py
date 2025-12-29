from django.shortcuts import render
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import status as http_status
import threading
import time 
import math
import requests
import json


class ComputeAsyncView(APIView):
    """Принимает данные расчета выбросов и списка устройств, запускает фоновую задачу."""
    
    def post(self, request):
        print("\n" + "="*50)
        print("[Django] Получен запрос на расчет")
        
        payload = request.data
        print(f"[Django] Данные запроса: {json.dumps(payload, indent=2)}")
        
        callback_url = payload.get('callback_url')
        callback_token = payload.get('callback_token')

        # ИСПРАВЛЕНО: Используем order_id, а не emission_id
        order_id = payload.get('order_id')
        distance = payload.get('distance')
        status_value = payload.get('status')  # переименовали чтобы не конфликтовать
        moderator_id = payload.get('moderator_id')
        devices = payload.get('devices', [])
        
        print(f"[Django] order_id: {order_id}")
        print(f"[Django] distance: {distance}")
        print(f"[Django] devices: {len(devices)} шт.")
        for i, device in enumerate(devices):
            print(f"  Device {i}: {device}")

        def worker():
            print(f"\n[Django Worker] Начало расчета для order_id={order_id}")
            time.sleep(5)  

            results = []
            total_emission = 0.0
            
            if status_value == "отклонен":
                print("[Django Worker] Статус 'отклонен', возвращаем нули")
                for device in devices:
                    device_id = device.get('device_id', 0)
                    results.append({
                        'device_id': device_id,
                        'current_emission': 0.0,
                        'power': 0.0
                    })
            else:
                print(f"[Django Worker] Расчет для {len(devices)} устройств")
                
                for device in devices:
                    try:
                        device_id = int(device.get('device_id', 0))
                        power = device.get('power')
                        
                        print(f"[Django Worker] Device {device_id}: power={power}, distance={distance}")
                        
                        if power is None or distance is None:
                            print(f"[Django Worker] Пропуск device {device_id}: нет power или distance")
                            results.append({
                                'device_id': device_id,
                                'current_emission': 0.0,
                                'power': 0.0,
                            })
                            continue

                        power_f = float(power)
                        distance_f = float(distance)
                        
                        if distance_f == 0:
                            print(f"[Django Worker] distance=0 для device {device_id}")
                            current_emission = 0.0
                        else:
                            current_emission = power_f / ((abs(math.modf(distance_f)[0]) * abs(math.modf(distance_f)[1])) * (abs(math.modf(distance_f)[0]) * abs(math.modf(distance_f)[1])))
                        
                        total_emission += current_emission
                        
                        print(f"[Django Worker] Device {device_id}: current_emission={current_emission}")
                        
                        results.append({
                            'device_id': device_id,
                            'current_emission': current_emission,
                            'power': power_f,
                        })
                        
                    except Exception as e:
                        print(f"[Django Worker] Ошибка для device {device.get('device_id')}: {e}")
                        results.append({
                            'device_id': device.get('device_id', 0),
                            'current_emission': 0.0,
                            'power': 0.0,
                        })
            
            # ИСПРАВЛЕНО: Правильная структура ответа
            response_data = {
                'results': results,
                'total_emission': total_emission,  # отдельное поле, не внутри results
                'order_id': order_id,  # ИСПРАВЛЕНО: используем order_id
                'status': 'completed' if status_value != "отклонен" else 'rejected',
                'moderator_id': moderator_id
            }
            
            print(f"\n[Django Worker] Итоговые данные:")
            print(f"  total_emission: {total_emission}")
            print(f"  order_id: {order_id}")
            print(f"  results: {len(results)} записей")
            for r in results:
                print(f"    device_id={r['device_id']}, emission={r['current_emission']}, power={r['power']}")
            
            # Отправка результатов на callback
            if callback_url:
                print(f"\n[Django Worker] Отправка на callback: {callback_url}")
                print(f"Данные: {json.dumps(response_data, indent=2)}")
                
                try:
                    # ИСПРАВЛЕНО: Отправляем правильную структуру
                    resp = requests.put(
                        callback_url, 
                        json=response_data, 
                        headers={'X-Calc-Token': str(callback_token)}, 
                        timeout=10
                    )
                    print(f"[Django Worker] Ответ от Go: статус={resp.status_code}")
                    print(f"[Django Worker] Тело ответа: {resp.text}")
                except Exception as e:
                    print(f"[Django Worker] Ошибка отправки callback: {e}")
            else:
                print("[Django Worker] Нет callback_url, отправка пропущена")
            
            print("[Django Worker] Расчет завершен")
            print("="*50 + "\n")

        # Запускаем в фоне
        thread = threading.Thread(target=worker, daemon=True)
        thread.start()
        
        print("[Django] Фоновый поток запущен, возвращаем ответ клиенту")

        return Response({
            'status': 'accepted', 
            'order_id': order_id,  # ИСПРАВЛЕНО
            'message': 'Расчет нагрузки запущен',
            'devices_count': len(devices)
        }, status=http_status.HTTP_200_OK)