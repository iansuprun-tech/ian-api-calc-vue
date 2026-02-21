<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const CURRENCY_NAMES: Record<string, string> = {
  AED: 'Дирхам ОАЭ',
  AFN: 'Афгани',
  ALL: 'Албанский лек',
  AMD: 'Армянский драм',
  ANG: 'Нидерландский антильский гульден',
  AOA: 'Ангольская кванза',
  ARS: 'Аргентинское песо',
  AUD: 'Австралийский доллар',
  AWG: 'Арубанский флорин',
  AZN: 'Азербайджанский манат',
  BAM: 'Конвертируемая марка',
  BBD: 'Барбадосский доллар',
  BDT: 'Бангладешская така',
  BGN: 'Болгарский лев',
  BHD: 'Бахрейнский динар',
  BIF: 'Бурундийский франк',
  BMD: 'Бермудский доллар',
  BND: 'Брунейский доллар',
  BOB: 'Боливиано',
  BRL: 'Бразильский реал',
  BSD: 'Багамский доллар',
  BTN: 'Бутанский нгултрум',
  BWP: 'Ботсванская пула',
  BYN: 'Белорусский рубль',
  BZD: 'Белизский доллар',
  CAD: 'Канадский доллар',
  CDF: 'Конголезский франк',
  CHF: 'Швейцарский франк',
  CLP: 'Чилийское песо',
  CNY: 'Китайский юань',
  COP: 'Колумбийское песо',
  CRC: 'Костариканский колон',
  CUP: 'Кубинское песо',
  CVE: 'Эскудо Кабо-Верде',
  CZK: 'Чешская крона',
  DJF: 'Франк Джибути',
  DKK: 'Датская крона',
  DOP: 'Доминиканское песо',
  DZD: 'Алжирский динар',
  EGP: 'Египетский фунт',
  ERN: 'Эритрейская накфа',
  ETB: 'Эфиопский быр',
  EUR: 'Евро',
  FJD: 'Фиджийский доллар',
  FKP: 'Фунт Фолклендских островов',
  FOK: 'Фарерская крона',
  GBP: 'Фунт стерлингов',
  GEL: 'Грузинский лари',
  GGP: 'Гернсийский фунт',
  GHS: 'Ганский седи',
  GIP: 'Гибралтарский фунт',
  GMD: 'Гамбийский даласи',
  GNF: 'Гвинейский франк',
  GTQ: 'Гватемальский кетсаль',
  GYD: 'Гайанский доллар',
  HKD: 'Гонконгский доллар',
  HNL: 'Гондурасская лемпира',
  HRK: 'Хорватская куна',
  HTG: 'Гаитянский гурд',
  HUF: 'Венгерский форинт',
  IDR: 'Индонезийская рупия',
  ILS: 'Новый израильский шекель',
  IMP: 'Фунт острова Мэн',
  INR: 'Индийская рупия',
  IQD: 'Иракский динар',
  IRR: 'Иранский риал',
  ISK: 'Исландская крона',
  JEP: 'Джерсийский фунт',
  JMD: 'Ямайский доллар',
  JOD: 'Иорданский динар',
  JPY: 'Японская иена',
  KES: 'Кенийский шиллинг',
  KGS: 'Киргизский сом',
  KHR: 'Камбоджийский риель',
  KID: 'Доллар Кирибати',
  KMF: 'Коморский франк',
  KRW: 'Южнокорейская вона',
  KWD: 'Кувейтский динар',
  KYD: 'Доллар Каймановых островов',
  KZT: 'Казахстанский тенге',
  LAK: 'Лаосский кип',
  LBP: 'Ливанский фунт',
  LKR: 'Шри-ланкийская рупия',
  LRD: 'Либерийский доллар',
  LSL: 'Лесотский лоти',
  LYD: 'Ливийский динар',
  MAD: 'Марокканский дирхам',
  MDL: 'Молдавский лей',
  MGA: 'Малагасийский ариари',
  MKD: 'Македонский денар',
  MMK: 'Мьянманский кьят',
  MNT: 'Монгольский тугрик',
  MOP: 'Патака Макао',
  MRU: 'Мавританская угия',
  MUR: 'Маврикийская рупия',
  MVR: 'Мальдивская руфия',
  MWK: 'Малавийская квача',
  MXN: 'Мексиканское песо',
  MYR: 'Малайзийский ринггит',
  MZN: 'Мозамбикский метикал',
  NAD: 'Намибийский доллар',
  NGN: 'Нигерийская найра',
  NIO: 'Никарагуанская кордоба',
  NOK: 'Норвежская крона',
  NPR: 'Непальская рупия',
  NZD: 'Новозеландский доллар',
  OMR: 'Оманский риал',
  PAB: 'Панамский бальбоа',
  PEN: 'Перуанский соль',
  PGK: 'Кина Папуа — Новой Гвинеи',
  PHP: 'Филиппинское песо',
  PKR: 'Пакистанская рупия',
  PLN: 'Польский злотый',
  PYG: 'Парагвайский гуарани',
  QAR: 'Катарский риал',
  RON: 'Румынский лей',
  RSD: 'Сербский динар',
  RUB: 'Российский рубль',
  RWF: 'Руандийский франк',
  SAR: 'Саудовский риял',
  SBD: 'Доллар Соломоновых Островов',
  SCR: 'Сейшельская рупия',
  SDG: 'Суданский фунт',
  SEK: 'Шведская крона',
  SGD: 'Сингапурский доллар',
  SHP: 'Фунт Святой Елены',
  SLE: 'Сьерра-леонский леоне',
  SOS: 'Сомалийский шиллинг',
  SRD: 'Суринамский доллар',
  SSP: 'Южносуданский фунт',
  STN: 'Добра Сан-Томе и Принсипи',
  SYP: 'Сирийский фунт',
  SZL: 'Свазилендский лилангени',
  THB: 'Тайский бат',
  TJS: 'Таджикский сомони',
  TMT: 'Туркменский манат',
  TND: 'Тунисский динар',
  TOP: 'Тонганская паанга',
  TRY: 'Турецкая лира',
  TTD: 'Доллар Тринидада и Тобаго',
  TVD: 'Доллар Тувалу',
  TWD: 'Новый тайваньский доллар',
  TZS: 'Танзанийский шиллинг',
  UAH: 'Украинская гривна',
  UGX: 'Угандийский шиллинг',
  USD: 'Доллар США',
  UYU: 'Уругвайское песо',
  UZS: 'Узбекский сум',
  VES: 'Венесуэльский боливар',
  VND: 'Вьетнамский донг',
  VUV: 'Вануатский вату',
  WST: 'Самоанская тала',
  XAF: 'Франк КФА BEAC',
  XCD: 'Восточнокарибский доллар',
  XDR: 'Специальные права заимствования',
  XOF: 'Франк КФА BCEAO',
  XPF: 'Франк КФП',
  YER: 'Йеменский риал',
  ZAR: 'Южноафриканский рэнд',
  ZMW: 'Замбийская квача',
  ZWL: 'Зимбабвийский доллар',
}

const props = defineProps<{
  modelValue: string
  currencies: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const search = ref('')
const open = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const dropdownRef = ref<HTMLDivElement | null>(null)

watch(() => props.modelValue, (val) => {
  if (!open.value) {
    search.value = val
  }
})

const filteredCurrencies = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return props.currencies
  return props.currencies.filter((code) => {
    const name = CURRENCY_NAMES[code] ?? ''
    return code.toLowerCase().includes(q) || name.toLowerCase().includes(q)
  })
})

function getName(code: string): string {
  return CURRENCY_NAMES[code] ?? code
}

function onFocus() {
  open.value = true
  search.value = ''
}

function selectCurrency(code: string) {
  emit('update:modelValue', code)
  search.value = code
  open.value = false
  inputRef.value?.blur()
}

function onBlur(e: FocusEvent) {
  const related = e.relatedTarget as HTMLElement | null
  if (dropdownRef.value?.contains(related)) return
  open.value = false
  search.value = props.modelValue
}
</script>

<template>
  <div class="currency-picker">
    <input
      ref="inputRef"
      :value="open ? search : modelValue"
      @input="search = ($event.target as HTMLInputElement).value"
      @focus="onFocus"
      @blur="onBlur"
      placeholder="Валюта"
      class="input-field"
      autocomplete="off"
    />
    <div v-if="open && filteredCurrencies.length > 0" ref="dropdownRef" class="dropdown" tabindex="-1">
      <div
        v-for="code in filteredCurrencies"
        :key="code"
        class="dropdown-item"
        @mousedown.prevent="selectCurrency(code)"
      >
        <span class="dropdown-code">{{ code }}</span>
        <span class="dropdown-name">{{ getName(code) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.currency-picker {
  position: relative;
}

.input-field {
  padding: 0.65rem 0.9rem;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: border-color 0.2s;
  background: #fff;
  color: #333;
  width: 130px;
}

.input-field:focus {
  outline: none;
  border-color: #0f3460;
}

.dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 300px;
  max-height: 260px;
  overflow-y: auto;
  background: #fff;
  border: 1.5px solid #ddd;
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  z-index: 100;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.55rem 0.9rem;
  cursor: pointer;
  transition: background 0.1s;
}

.dropdown-item:first-child {
  border-radius: 10px 10px 0 0;
}

.dropdown-item:last-child {
  border-radius: 0 0 10px 10px;
}

.dropdown-item:hover {
  background: #f0f4ff;
}

.dropdown-code {
  font-weight: 600;
  font-size: 0.9rem;
  color: #0f3460;
  min-width: 36px;
}

.dropdown-name {
  font-size: 0.85rem;
  color: #666;
}
</style>
