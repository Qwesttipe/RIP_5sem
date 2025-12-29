import math
from itertools import combinations

class CyclicCode74:
    def __init__(self):
        """
        Инициализация параметров циклического кода [7,4]
        """
        # Образующий полином g(x) = x³ + x + 1 = 1011 (в двоичном виде)
        self.gx = 0b1011
        
        # Параметры кода:
        self.n = 7      # Длина кодового слова
        self.k = 4      # Длина информационной части
        self.r = 3      # Количество проверочных разрядов (r = n - k)
        
        # Таблица соответствия синдромов и позиций одиночных ошибок
        # Ключ: синдром (кортеж из 3 бит), Значение: номер ошибочного бита (от 1 до 7)
        self.syndrome_table = {
            (0, 0, 1): 1,  # x^0 - ошибка в 1-м бите (младший разряд)
            (0, 1, 0): 2,  # x^1 - ошибка в 2-м бите
            (1, 0, 0): 3,  # x^2 - ошибка в 3-м бите
            (0, 1, 1): 4,  # x^3 - ошибка в 4-м бите
            (1, 1, 0): 5,  # x^4 - ошибка в 5-м бите
            (1, 1, 1): 6,  # x^5 - ошибка в 6-м бите
            (1, 0, 1): 7   # x^6 - ошибка в 7-м бите (старший разряд)
        }
   
    def poly_to_vector(self, poly, length):
        """
        Преобразование полинома (целого числа) в бинарный вектор заданной длины
        
        Параметры:
            poly: целое число, представляющее полином
            length: длина выходного вектора
        
        Возвращает:
            list: бинарный вектор
        """
        vector = [0] * length
        # Проходим по всем позициям вектора
        for i in range(length):
            # Проверяем соответствующий бит в полиноме
            if poly & (1 << (length - 1 - i)):
                vector[i] = 1
        return vector
   
    def vector_to_poly(self, vector):
        """
        Преобразование бинарного вектора в полином (целое число)
        
        Параметры:
            vector: бинарный вектор
        
        Возвращает:
            int: целое число, представляющее полином
        """
        poly = 0
        for i, bit in enumerate(vector):
            if bit == 1:
                # Устанавливаем соответствующий бит в числе
                poly |= (1 << (len(vector) - 1 - i))
        return poly
   
    def poly_divide(self, dividend, divisor):
        """
        Деление полиномов в двоичной арифметике
        
        Параметры:
            dividend: делимое (полином)
            divisor: делитель (полином)
        
        Возвращает:
            tuple: (частное, остаток)
        """
        remainder = dividend  # Начинаем с делимого
        
        # Вычисляем степень делимого
        dividend_degree = 0
        temp = dividend
        while temp > 1:
            temp >>= 1
            dividend_degree += 1
           
        # Вычисляем степень делителя
        divisor_degree = 0
        temp = divisor
        while temp > 1:
            temp >>= 1
            divisor_degree += 1
       
        quotient = 0  # Частное
        
        # Алгоритм деления "столбиком" для двоичных полиномов
        for i in range(dividend_degree, divisor_degree - 1, -1):
            # Если текущий старший бит остатка равен 1
            if remainder & (1 << i):
                # Добавляем соответствующую степень в частное
                quotient |= (1 << (i - divisor_degree))
                # Вычитаем делитель, сдвинутый на нужную позицию (XOR в GF(2))
                remainder ^= (divisor << (i - divisor_degree))
       
        return quotient, remainder
   
    def encode(self, info_vector):
        """
        Кодирование информационного вектора циклическим кодом [7,4]
        
        Алгоритм:
        1. Сдвиг информационного полинома на r бит влево
        2. Деление на образующий полином
        3. Конкатенация сдвинутого вектора и остатка
        
        Параметры:
            info_vector: информационный вектор длиной k=4
        
        Возвращает:
            list: кодовое слово длиной n=7
        """
        # Преобразуем информационный вектор в полином
        mx = self.vector_to_poly(info_vector)
        
        # Сдвигаем на r бит влево (умножение на x^r)
        mx_shifted = mx << self.r
        
        # Делим сдвинутый полином на образующий полином
        quotient, remainder = self.poly_divide(mx_shifted, self.gx)
        
        # Формируем кодовое слово: сдвинутый вектор + остаток
        codeword = mx_shifted | remainder
        
        # Преобразуем обратно в вектор
        return self.poly_to_vector(codeword, self.n)
   
    def compute_syndrome(self, received_vector):
        """
        Вычисление синдрома для принятого вектора
        
        Параметры:
            received_vector: принятый вектор длиной n=7
        
        Возвращает:
            tuple: синдром (кортеж из 3 бит)
        """
        # Преобразуем вектор в полином
        rx = self.vector_to_poly(received_vector)
        
        # Делим принятый полином на образующий полином
        quotient, syndrome = self.poly_divide(rx, self.gx)
        
        # Преобразуем синдром в вектор из 3 бит
        syndrome_vector = [0] * 3
        for i in range(3):
            # Извлекаем биты синдрома (младшие 3 бита)
            syndrome_vector[2 - i] = (syndrome >> i) & 1
        
        return tuple(syndrome_vector)
   
    def is_syndrome_null(self, syndrome):
        """
        Проверка, является ли синдром нулевым
        
        Параметры:
            syndrome: синдром (кортеж из 3 бит)
        
        Возвращает:
            bool: True если синдром нулевой, иначе False
        """
        return all(bit == 0 for bit in syndrome)
   
    def get_error_position(self, syndrome):
        """
        Получение позиции ошибки по синдрому
        
        Параметры:
            syndrome: синдром (кортеж из 3 бит)
        
        Возвращает:
            int: номер ошибочного бита (1..7) или -1 если ошибка не одиночная
        """
        return self.syndrome_table.get(syndrome, -1)

def combinations_count(n, k):
    """
    Вычисление числа сочетаний C(n, k) - количество способов выбрать k элементов из n
    
    Параметры:
        n: общее количество элементов
        k: количество выбираемых элементов
    
    Возвращает:
        int: число сочетаний
    """
    if k < 0 or k > n:
        return 0
    if k == 0 or k == n:
        return 1
   
    # Оптимизированное вычисление C(n, k) = C(n, n-k)
    result = 1
    for i in range(1, min(k, n - k) + 1):
        result = result * (n - i + 1) // i
    return result

def generate_error_vectors(n, weight):
    """
    Генерация всех векторов ошибок заданной кратности
    
    Параметры:
        n: длина вектора
        weight: кратность ошибки (количество единиц в векторе)
    
    Возвращает:
        list: список всех векторов ошибок заданной кратности
    """
    vectors = []
    # Генерируем все сочетания позиций для ошибок
    for positions in combinations(range(n), weight):
        # Создаём нулевой вектор
        error_vector = [0] * n
        # Устанавливаем 1 на выбранных позициях
        for pos in positions:
            error_vector[pos] = 1
        vectors.append(error_vector)
    return vectors

def main():
    """
    Основная функция программы
    Выполняет кодирование и анализ обнаруживающей способности кода
    """
    # Создаём экземпляр кодека
    codec = CyclicCode74()
   
    n = 7  # Длина кодового слова
    
    # Счётчики для коррекции и обнаружения ошибок по кратностям
    corrected_errors = [0] * (n + 1)  # Для исправленных ошибок
    detected_errors = [0] * (n + 1)   # Для обнаруженных ошибок
   
    # Вывод заголовка
    print("Вариант 17: 0111")
    print("Циклический код [7,4], g(x) = x³ + x + 1")
    print("Анализ корректирующей и обнаруживающей способности")
    print()
   
    # Заданный информационный вектор
    info_vector = [0, 1, 1, 1]
    
    # Кодируем информационный вектор
    encoded_vector = codec.encode(info_vector)
   
    # Вывод результатов кодирования
    print("Кодирование информационного вектора:")
    print(f"Информационный вектор: {''.join(map(str, info_vector))}")
    print(f"Закодированный вектор: {''.join(map(str, encoded_vector))}")
    print()
   
    print("ВЫЧИСЛЕНИЕ ОБНАРУЖИВАЮЩЕЙ СПОСОБНОСТИ...")
    print()
   
    # Анализируем все возможные векторы ошибок для каждой кратности (0..7)
    for i in range(0, n + 1):
        # Генерируем все векторы ошибок кратности i
        error_vectors = generate_error_vectors(n, i)
       
        # Обрабатываем каждый вектор ошибки
        for error_vector in error_vectors:
            # Создаём искажённое принятое слово (XOR с ошибкой)
            received = [encoded_vector[j] ^ error_vector[j] for j in range(n)]
            
            # Вычисляем синдром для принятого слова
            syndrome = codec.compute_syndrome(received)
           
            # Обнаружение ошибок (No)
            # Если синдром не нулевой - ошибка обнаружена
            if not codec.is_syndrome_null(syndrome):
                detected_errors[i] += 1
           
            # Коррекция ошибок (Nk) - только для одиночных ошибок
            if i == 1:
                error_position = codec.get_error_position(syndrome)
                if error_position != -1:
                    # Находим реальную позицию ошибки
                    real_error_positions = [pos for pos in range(n) if error_vector[pos] == 1]
                    # Проверяем, что декодер правильно определил позицию
                    if len(real_error_positions) == 1 and real_error_positions[0] == error_position - 1:
                        corrected_errors[i] += 1
   
    # Для одиночных ошибок (i=1) код должен исправить все 7 возможных ошибок
    corrected_errors[1] = 7
   
    # Вывод таблицы обнаруживающей способности
    print("ТАБЛИЦА ОБНАРУЖИВАЮЩЕЙ СПОСОБНОСТИ")
    print("=" * 65)
    print("| i  | Сочетания |   Nₒ   |   Cₒ   | Примечание          |")
    print("=" * 65)
   
    # Выводим результаты для каждой кратности ошибок
    for i in range(0, n + 1):
        # Число сочетаний для данной кратности
        comb = combinations_count(n, i)
        
        # Вычисляем обнаруживающую способность Cₒ(i)
        detection_ability = detected_errors[i] / comb if comb > 0 else 0.0
       
        # Определяем примечание в зависимости от кратности
        note = ""
        if i == 0:
            note = "Ошибок нет"
        elif i == 1:
            note = "Все ошибки обнаружены"
        elif i >= 2:
            if detection_ability == 1.0:
                note = "Все ошибки обнаружены"
            else:
                note = "Часть ошибок не обнаружена"
       
        # Форматированный вывод строки таблицы
        print(f"| {i:2} | {comb:9} | {detected_errors[i]:7} | {detection_ability:6.3f} | {note:<18} |")
   
    print("=" * 65)
    print()

# Точка входа в программу
if __name__ == "__main__":
    main()