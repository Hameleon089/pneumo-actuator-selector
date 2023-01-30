# Pneumo actuator selector

Консольное приложение для подбора пневмоприводов [ArTorq](https://artorq.ru/ "Сайт производителя")

***

## Используемые технологии
* Go (1.19)

***

## Описание 

Приложение подбирает подходящий пневматический привод по заданным параметрам:

* номинальный крутящий момент запорной арматуры 
* рабочее давление воздуха
* необходимый коэффициент запаса

### Режимы работы приложения

* Подбор двухстороннего привода
* Подбор НЗ (нормально закрытого) привода с пружинами для затвора
* Подбор НЗ привода с пружинами для крана
* Подбор НЗ привода с ручным вводом всех коэффициентов запаса (BTO, ETO, BTC, ETC)

***

## Установка и запуск

1. Скопировать все содержимое репозитория
2. Скомпилировать приложение под вашу ОС (`go build ...`, для Windows в репозитории уже имеется скомпилированное приложение - файл `pneumo-actuator-selector.exe`)
3. Запустить скомпилированный файл приложения 

***

## Примеры работы приложения 

### Вид основного меню
    Подбор пневмоприводов ArTorq

    ВНИМАНИЕ! Нецелые числа необходимо писать ЧЕРЕЗ ТОЧКУ.
    Например: 1.25 - верно; 1,25 - неверно.
    
    Выберите режим работы программы:
    1 - подбор двухстороннего привода
    2 - подбор НЗ привода с пружинами для затвора
    3 - подбор НЗ привода с пружинами для крана
    4 - подбор НЗ привода с ручным вводом всех коэффициентов запаса (BTO, ETO, BTC, ETC)
    5 - завершить работу программы
    Введите число (1-5): 

### Подбор двухстороннего привода
    ...основное меню...
    Введите число (1-5): 1
    Введите номинальный крутящий момент (Н*м): 120
    Введите рабочее давление (бар): 6
    Введите коэффициент запаса (например 1.25): 1.25
    
    Модель привода - PA16DA
    Коэффициент запаса - 1.33

### Подбор НЗ привода с пружинами для затвора
    ...основное меню...
    Введите число (1-5): 2
    Введите номинальный крутящий момент (Н*м): 646
    Введите рабочее давление (бар): 5
    Введите коэффициент запаса (например 1.25): 1.3
    
    Модель привода - PA220SR
    Пружины номер - 9
    Коэффициенты запаса:
    BTO - 1.58 (задан - 1.30)
    ETO - 1.18 (задан - 0.50)
    BTC - 1.72 (задан - 1.00)
    ETC - 1.32 (задан - 1.30)

### Подбор НЗ привода с ручным вводом всех коэффициентов запаса (BTO, ETO, BTC, ETC)
    ...основное меню...
    Введите число (1-5): 4
    Введите номинальный крутящий момент (Н*м): 646
    Введите рабочее давление (бар): 5
    Введите коэффициент запаса BTO (например 1.25): 1.25
    Введите коэффициент запаса ETO (например 0.5): 0.5
    Введите коэффициент запаса BTC (например 1.0): 1.0
    Введите коэффициент запаса ETC (например 1.25): 1.25
    
    Модель привода - PA220SR
    Пружины номер - 9
    Коэффициенты запаса:
    BTO - 1.58 (задан - 1.25)
    ETO - 1.18 (задан - 0.50)
    BTC - 1.72 (задан - 1.00)
    ETC - 1.32 (задан - 1.25)